package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 行为类型常量
const (
	BehaviorView     = int64(1) // 浏览
	BehaviorClick    = int64(2) // 点击
	BehaviorAddCart  = int64(3) // 加购
	BehaviorPurchase = int64(4) // 购买
	BehaviorCollect  = int64(5) // 收藏
)

// 行为权重（用于推荐评分）
var BehaviorWeight = map[int64]float64{
	BehaviorView:     1.0,
	BehaviorClick:    2.0,
	BehaviorAddCart:  3.0,
	BehaviorCollect:  4.0,
	BehaviorPurchase: 5.0,
}

var _ UserBehaviorModel = (*customUserBehaviorModel)(nil)

type (
	UserBehaviorModel interface {
		userBehaviorModel
		// 查询用户最近的行为记录
		FindRecentByUserId(ctx context.Context, userId int64, limit int) ([]*UserBehavior, error)
		// 查询用户对某些商品的行为（去重商品维度，取最高权重行为）
		FindUserProductBehaviors(ctx context.Context, userId int64, days int) ([]*UserProductScore, error)
		// 查询与某用户行为相似的其他用户（UserCF: 共同交互商品数）
		FindSimilarUsers(ctx context.Context, userId int64, days int, limit int) ([]int64, error)
		// 查询某些用户交互过的商品（排除指定商品）
		FindProductsByUsers(ctx context.Context, userIds []int64, excludeProductIds []int64, days int, limit int) ([]*UserBehavior, error)
		// 批量插入行为
		BatchInsert(ctx context.Context, behaviors []*UserBehavior) error
		// 查询商品的行为用户列表（ItemCF 用）
		FindUsersByProduct(ctx context.Context, productId int64, days int) ([]int64, error)
		// 查询用户最近交互的商品ID列表
		FindRecentProductIds(ctx context.Context, userId int64, limit int) ([]int64, error)
		// 查询用户偏好的分类（按行为权重排序）
		FindUserPreferredCategories(ctx context.Context, userId int64, days int, limit int) ([]int64, error)
	}

	// UserProductScore 用户对商品的综合行为评分
	UserProductScore struct {
		ProductId  int64   `db:"product_id"`
		CategoryId int64   `db:"category_id"`
		Score      float64 `db:"score"`
	}

	customUserBehaviorModel struct {
		*defaultUserBehaviorModel
	}
)

func NewUserBehaviorModel(conn sqlx.SqlConn) UserBehaviorModel {
	return &customUserBehaviorModel{
		defaultUserBehaviorModel: newUserBehaviorModel(conn),
	}
}

func (m *customUserBehaviorModel) FindRecentByUserId(ctx context.Context, userId int64, limit int) ([]*UserBehavior, error) {
	query := fmt.Sprintf("select %s from %s where `user_id`=? order by `behavior_time` desc limit ?", userBehaviorRows, m.table)
	var resp []*UserBehavior
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, limit)
	return resp, err
}

// FindUserProductBehaviors 获取用户近N天内对各商品的加权行为评分
// 返回按评分降序排列的商品列表
func (m *customUserBehaviorModel) FindUserProductBehaviors(ctx context.Context, userId int64, days int) ([]*UserProductScore, error) {
	query := fmt.Sprintf(`
		SELECT product_id, category_id,
			SUM(CASE behavior_type
				WHEN 1 THEN 1.0
				WHEN 2 THEN 2.0
				WHEN 3 THEN 3.0
				WHEN 4 THEN 5.0
				WHEN 5 THEN 4.0
				ELSE 0
			END) as score
		FROM %s
		WHERE user_id = ? AND behavior_time > ?
		GROUP BY product_id, category_id
		ORDER BY score DESC`, m.table)
	cutoff := nowMs() - int64(days)*86400*1000
	var resp []*UserProductScore
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, cutoff)
	return resp, err
}

// FindSimilarUsers 基于共同交互商品找相似用户（简化版 UserCF）
func (m *customUserBehaviorModel) FindSimilarUsers(ctx context.Context, userId int64, days int, limit int) ([]int64, error) {
	cutoff := nowMs() - int64(days)*86400*1000
	query := fmt.Sprintf(`
		SELECT b2.user_id
		FROM %s b1
		JOIN %s b2 ON b1.product_id = b2.product_id AND b2.user_id != b1.user_id
		WHERE b1.user_id = ? AND b1.behavior_time > ? AND b2.behavior_time > ?
		GROUP BY b2.user_id
		ORDER BY COUNT(DISTINCT b2.product_id) DESC
		LIMIT ?`, m.table, m.table)
	var resp []int64
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, cutoff, cutoff, limit)
	return resp, err
}

// FindProductsByUsers 查询指定用户群交互过的商品（排除已知商品）
func (m *customUserBehaviorModel) FindProductsByUsers(ctx context.Context, userIds []int64, excludeProductIds []int64, days int, limit int) ([]*UserBehavior, error) {
	if len(userIds) == 0 {
		return nil, nil
	}
	cutoff := nowMs() - int64(days)*86400*1000

	userPlaceholders := strings.Repeat("?,", len(userIds))
	userPlaceholders = userPlaceholders[:len(userPlaceholders)-1]

	args := make([]interface{}, 0, len(userIds)+len(excludeProductIds)+2)
	for _, uid := range userIds {
		args = append(args, uid)
	}
	args = append(args, cutoff)

	excludeClause := ""
	if len(excludeProductIds) > 0 {
		excludePlaceholders := strings.Repeat("?,", len(excludeProductIds))
		excludePlaceholders = excludePlaceholders[:len(excludePlaceholders)-1]
		excludeClause = fmt.Sprintf("AND product_id NOT IN (%s)", excludePlaceholders)
		for _, pid := range excludeProductIds {
			args = append(args, pid)
		}
	}
	args = append(args, limit)

	query := fmt.Sprintf(`
		SELECT product_id, category_id, MAX(behavior_type) as behavior_type, MAX(behavior_time) as behavior_time, 0 as id, 0 as user_id
		FROM %s
		WHERE user_id IN (%s) AND behavior_time > ? %s
		GROUP BY product_id, category_id
		ORDER BY COUNT(*) DESC
		LIMIT ?`, m.table, userPlaceholders, excludeClause)

	var resp []*UserBehavior
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	return resp, err
}

func (m *customUserBehaviorModel) BatchInsert(ctx context.Context, behaviors []*UserBehavior) error {
	if len(behaviors) == 0 {
		return nil
	}
	values := make([]string, 0, len(behaviors))
	args := make([]interface{}, 0, len(behaviors)*5)
	for _, b := range behaviors {
		values = append(values, "(?, ?, ?, ?, ?)")
		args = append(args, b.UserId, b.ProductId, b.CategoryId, b.BehaviorType, b.BehaviorTime)
	}
	query := fmt.Sprintf("insert into %s (%s) values %s", m.table, userBehaviorRowsExpectAutoSet, strings.Join(values, ","))
	_, err := m.conn.ExecCtx(ctx, query, args...)
	return err
}

func (m *customUserBehaviorModel) FindUsersByProduct(ctx context.Context, productId int64, days int) ([]int64, error) {
	cutoff := nowMs() - int64(days)*86400*1000
	query := fmt.Sprintf("select distinct `user_id` from %s where `product_id`=? and `behavior_time`>?", m.table)
	var resp []int64
	err := m.conn.QueryRowsCtx(ctx, &resp, query, productId, cutoff)
	return resp, err
}

func (m *customUserBehaviorModel) FindRecentProductIds(ctx context.Context, userId int64, limit int) ([]int64, error) {
	query := fmt.Sprintf(`
		SELECT product_id FROM (
			SELECT product_id, MAX(behavior_time) as latest
			FROM %s WHERE user_id=?
			GROUP BY product_id
			ORDER BY latest DESC
			LIMIT ?
		) t`, m.table)
	var resp []int64
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, limit)
	return resp, err
}

// FindUserPreferredCategories 获取用户偏好分类（按加权行为评分排序）
func (m *customUserBehaviorModel) FindUserPreferredCategories(ctx context.Context, userId int64, days int, limit int) ([]int64, error) {
	cutoff := nowMs() - int64(days)*86400*1000
	query := fmt.Sprintf(`
		SELECT category_id
		FROM %s
		WHERE user_id = ? AND behavior_time > ?
		GROUP BY category_id
		ORDER BY SUM(CASE behavior_type
			WHEN 1 THEN 1 WHEN 2 THEN 2 WHEN 3 THEN 3 WHEN 4 THEN 5 WHEN 5 THEN 4 ELSE 0
		END) DESC
		LIMIT ?`, m.table)
	var resp []int64
	err := m.conn.QueryRowsCtx(ctx, &resp, query, userId, cutoff, limit)
	return resp, err
}

func nowMs() int64 {
	return timeNow().UnixMilli()
}
