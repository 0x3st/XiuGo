package bbs

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/0x3st/XiuGo/internal/dao"
	"github.com/0x3st/XiuGo/internal/model/do"
	"github.com/0x3st/XiuGo/internal/model/entity"
)

// SyncPHPRuntime keeps the cached counters used by Xiuno's PHP frontend in
// step with mutations performed by Go. Xiuno's own InnoDB fallback reads the
// approximate information_schema TABLE_ROWS value, so deleting the cache is
// not sufficient after a write.
func (s *Service) SyncPHPRuntime(ctx context.Context, todayChanges map[string]int) error {
	runtime, cacheKey, exists, err := s.loadPHPRuntime(ctx)
	if err != nil {
		return err
	}
	if runtime["users"], err = dao.BbsUser.Ctx(ctx).Count(); err != nil {
		return gerror.Wrap(err, "同步 PHP 用户统计失败")
	}
	if runtime["threads"], err = dao.BbsThread.Ctx(ctx).Count(); err != nil {
		return gerror.Wrap(err, "同步 PHP 主题统计失败")
	}
	if runtime["posts"], err = dao.BbsPost.Ctx(ctx).Where(do.BbsPost{Isfirst: 0}).Count(); err != nil {
		return gerror.Wrap(err, "同步 PHP 回复统计失败")
	}
	onlineHoldSeconds := s.originalOnlineHoldSeconds(ctx)
	if runtime["onlines"], err = dao.BbsSession.Ctx(ctx).
		WhereGT(dao.BbsSession.Columns().LastDate, uint(time.Now().Unix()-int64(onlineHoldSeconds))).Count(); err != nil {
		return gerror.Wrap(err, "同步 PHP 在线统计失败")
	}
	if runtime["onlines"] < 1 {
		runtime["onlines"] = 1
	}
	for _, key := range []string{"todayusers", "todaythreads", "todayposts", "cron_1_last_date", "cron_2_last_date"} {
		if _, found := runtime[key]; !found {
			runtime[key] = 0
		}
	}
	for key, change := range todayChanges {
		runtime[key] += change
		if runtime[key] < 0 {
			runtime[key] = 0
		}
	}
	return s.savePHPRuntime(ctx, cacheKey, exists, runtime)
}

func (s *Service) loadPHPRuntime(ctx context.Context) (runtime map[string]int, cacheKey string, exists bool, err error) {
	cacheKeys := s.phpCacheKeys(ctx, "runtime")
	cacheKey = cacheKeys[len(cacheKeys)-1]
	var cached entity.BbsCache
	for _, candidate := range cacheKeys {
		cached = entity.BbsCache{}
		if err = dao.BbsCache.Ctx(ctx).Where(do.BbsCache{K: candidate}).Scan(&cached); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, "", false, gerror.Wrap(err, "读取 PHP runtime 缓存失败")
		}
		if cached.K != "" {
			cacheKey = candidate
			exists = true
			break
		}
	}
	runtime = make(map[string]int)
	if cached.V != "" {
		if err = json.Unmarshal([]byte(cached.V), &runtime); err != nil {
			return nil, "", false, gerror.Wrap(err, "解析 PHP runtime 缓存失败")
		}
	}
	return runtime, cacheKey, exists, nil
}

func (s *Service) savePHPRuntime(
	ctx context.Context, cacheKey string, exists bool, runtime map[string]int,
) error {
	encoded, err := json.MarshalIndent(runtime, "", "    ")
	if err != nil {
		return gerror.Wrap(err, "编码 PHP runtime 缓存失败")
	}
	if !exists {
		if _, err = dao.BbsCache.Ctx(ctx).Data(do.BbsCache{K: cacheKey, V: string(encoded), Expiry: 0}).Insert(); err != nil {
			return gerror.Wrap(err, "创建 PHP runtime 缓存失败")
		}
		return nil
	}
	if _, err = dao.BbsCache.Ctx(ctx).Where(do.BbsCache{K: cacheKey}).
		Data(do.BbsCache{V: string(encoded), Expiry: 0}).Update(); err != nil {
		return gerror.Wrap(err, "更新 PHP runtime 缓存失败")
	}
	return nil
}

func (s *Service) originalRuntimeOnlines(ctx context.Context) (int, error) {
	runtime, _, _, err := s.loadPHPRuntime(ctx)
	if err != nil {
		return 0, err
	}
	onlines := runtime["onlines"]
	if onlines < 1 {
		onlines = 1
	}
	return onlines, nil
}
