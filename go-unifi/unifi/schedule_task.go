package unifi

import (
	"context"
)

func (c *client) ListScheduleTask(ctx context.Context, site string) ([]ScheduleTask, error) {
	return c.listScheduleTask(ctx, site)
}

func (c *client) GetScheduleTask(ctx context.Context, site, id string) (*ScheduleTask, error) {
	return c.getScheduleTask(ctx, site, id)
}

func (c *client) CreateScheduleTask(ctx context.Context, site string, d *ScheduleTask) (*ScheduleTask, error) {
	return c.createScheduleTask(ctx, site, d)
}

func (c *client) UpdateScheduleTask(ctx context.Context, site string, d *ScheduleTask) (*ScheduleTask, error) {
	return c.updateScheduleTask(ctx, site, d)
}

func (c *client) DeleteScheduleTask(ctx context.Context, site, id string) error {
	return c.deleteScheduleTask(ctx, site, id)
}
