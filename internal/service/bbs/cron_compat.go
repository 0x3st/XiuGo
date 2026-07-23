package bbs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"

	"github.com/0x3st/XiuGo/internal/dao"
	"github.com/0x3st/XiuGo/internal/model/do"
)

const (
	fiveMinuteInterval = 5 * time.Minute
	dailyInterval      = 24 * time.Hour
)

func StartOriginalMaintenance(ctx context.Context) error {
	service := New()
	if err := service.RunDueMaintenance(ctx, time.Now(), false); err != nil {
		return err
	}
	_, err := gcron.AddSingleton(ctx, "@every 1m", func(jobCtx context.Context) {
		if runErr := service.RunDueMaintenance(jobCtx, time.Now(), false); runErr != nil {
			g.Log().Error(jobCtx, runErr)
		}
	}, "xiuno-original-maintenance")
	return gerror.Wrap(err, "注册 Xiuno 原版计划任务失败")
}

func (s *Service) RunDueMaintenance(ctx context.Context, now time.Time, force bool) error {
	runtime, _, _, err := s.loadPHPRuntime(ctx)
	if err != nil {
		return err
	}
	if len(runtime) == 0 {
		if err = s.SyncPHPRuntime(ctx, nil); err != nil {
			return err
		}
		runtime, _, _, err = s.loadPHPRuntime(ctx)
		if err != nil {
			return err
		}
	}
	nowUnix := now.Unix()
	if force || maintenanceDue(int64(runtime["cron_1_last_date"]), nowUnix, fiveMinuteInterval) {
		if err = s.runFiveMinuteMaintenance(ctx, now); err != nil {
			return err
		}
	}
	if force || maintenanceDue(int64(runtime["cron_2_last_date"]), nowUnix, dailyInterval) {
		if err = s.runDailyMaintenance(ctx, now); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) runFiveMinuteMaintenance(ctx context.Context, now time.Time) error {
	holdSeconds := s.originalOnlineHoldSeconds(ctx)
	expiry := uint(now.Unix() - int64(holdSeconds))
	if _, err := dao.BbsSession.Ctx(ctx).WhereLT(dao.BbsSession.Columns().LastDate, expiry).Delete(); err != nil {
		return gerror.Wrap(err, "清理过期 PHP Session 失败")
	}
	if _, err := dao.BbsSessionData.Ctx(ctx).WhereLT(dao.BbsSessionData.Columns().LastDate, expiry).Delete(); err != nil {
		return gerror.Wrap(err, "清理过期 PHP Session 数据失败")
	}
	if err := s.SyncPHPRuntime(ctx, nil); err != nil {
		return err
	}
	return s.setPHPRuntimeValues(ctx, map[string]int{"cron_1_last_date": int(now.Unix())})
}

func (s *Service) runDailyMaintenance(ctx context.Context, now time.Time) error {
	if _, err := dao.BbsForum.Ctx(ctx).WhereGT(dao.BbsForum.Columns().Fid, 0).Data(do.BbsForum{
		Todayposts: 0, Todaythreads: 0,
	}).Update(); err != nil {
		return gerror.Wrap(err, "重置板块今日统计失败")
	}
	if err := s.cleanupOriginalTempAttachments(ctx, now); err != nil {
		return err
	}
	if _, err := dao.BbsQueue.Ctx(ctx).WhereLT(dao.BbsQueue.Columns().Expiry, uint(now.Unix())).Delete(); err != nil {
		return gerror.Wrap(err, "清理过期主题队列失败")
	}
	if err := s.SyncPHPRuntime(ctx, nil); err != nil {
		return err
	}
	midnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return s.setPHPRuntimeValues(ctx, map[string]int{
		"todayposts": 0, "todaythreads": 0, "todayusers": 0,
		"cron_2_last_date": int(midnight.Unix()),
	})
}

func (s *Service) setPHPRuntimeValues(ctx context.Context, values map[string]int) error {
	runtime, cacheKey, exists, err := s.loadPHPRuntime(ctx)
	if err != nil {
		return err
	}
	for key, value := range values {
		runtime[key] = value
	}
	return s.savePHPRuntime(ctx, cacheKey, exists, runtime)
}

func (s *Service) originalOnlineHoldSeconds(ctx context.Context) int {
	if seconds := s.phpConfInt(ctx, "online_hold_time"); seconds > 0 {
		return seconds
	}
	if seconds := g.Cfg().MustGet(ctx, "xiuno.onlineHoldTime", 3600).Int(); seconds > 0 {
		return seconds
	}
	return 3600
}

func (s *Service) cleanupOriginalTempAttachments(ctx context.Context, now time.Time) error {
	directory := filepath.Join(s.uploadRoot(ctx), "tmp")
	entries, err := os.ReadDir(directory)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return gerror.Wrap(err, "读取临时附件目录失败")
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.Contains(entry.Name(), ".") {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			return gerror.Wrap(infoErr, "读取临时附件信息失败")
		}
		if now.Sub(info.ModTime()) > dailyInterval {
			if removeErr := os.Remove(filepath.Join(directory, entry.Name())); removeErr != nil && !os.IsNotExist(removeErr) {
				return gerror.Wrap(removeErr, "清理临时附件失败")
			}
		}
	}
	return nil
}

func maintenanceDue(lastUnix, nowUnix int64, interval time.Duration) bool {
	return nowUnix-lastUnix > int64(interval/time.Second)
}
