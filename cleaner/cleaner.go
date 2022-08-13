package cleaner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang/glog"
	"github.com/google/go-github/v45/github"
)

type Cleaner struct {
	client *github.Client
	// org is the organization to clean up runners for.
	org string
	// repository is the repository to clean up runners for.
	// if empty, organization runners will be cleaned
	repository        string
	offlineCountsByID map[int64]int
	offlineThreshold  int
}

func NewCleaner(client *github.Client, org string, offlineThreshold int) *Cleaner {
	return NewCleanerWithRepository(client, org, "", offlineThreshold)
}

func NewCleanerWithRepository(client *github.Client, org, repository string, offlineThreshold int) *Cleaner {
	return &Cleaner{
		client:            client,
		org:               org,
		repository:        repository,
		offlineCountsByID: make(map[int64]int),
		offlineThreshold:  offlineThreshold,
	}
}

func (c *Cleaner) Run(ctx context.Context) error {
	err := c.clean(ctx)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			err := c.clean(ctx)
			if err != nil {
				return err
			}
		}
	}
}

func (c *Cleaner) clean(ctx context.Context) error {
	if c.repository == "" {
		glog.Infof("Cleaning up runners for %s...", c.org)
	} else {
		glog.Infof("Cleaning up runners for %s/%s...", c.org, c.repository)
	}
	opts := &github.ListOptions{}

	offlineRunners := make(map[int64]struct{})
	for {
		var runners *github.Runners
		var resp *github.Response
		var err error
		if c.repository == "" {
			runners, resp, err = c.client.Actions.ListOrganizationRunners(ctx, c.org, opts)
			if err != nil {
				return err
			}
		} else {
			runners, resp, err = c.client.Actions.ListRunners(ctx, c.org, c.repository, opts)
			if err != nil {
				return err
			}
		}

		for _, runner := range runners.Runners {
			if *runner.Status == "offline" {
				offlineRunners[*runner.ID] = struct{}{}
			}
		}
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}
	glog.Infof("Found %d offline runners", len(offlineRunners))

	for id := range c.offlineCountsByID {
		if _, ok := offlineRunners[id]; !ok {
			delete(c.offlineCountsByID, id)
		}
	}

	for id := range offlineRunners {
		c.offlineCountsByID[id]++
		if c.offlineCountsByID[id] < c.offlineThreshold {
			continue
		}

		var resp *github.Response
		var err error
		if c.repository == "" {
			resp, err = c.client.Actions.RemoveOrganizationRunner(ctx, c.org, id)
			if err != nil {
				return err
			}
		} else {
			resp, err = c.client.Actions.RemoveRunner(ctx, c.org, c.repository, id)
			if err != nil {
				return err
			}
		}

		if resp.StatusCode != http.StatusNoContent {
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			return fmt.Errorf("failed to remove runner code: %d: %s", resp.StatusCode, body)
		}

		delete(c.offlineCountsByID, id)
		if c.repository == "" {
			glog.Infof("Removed runner %d from %s", id, c.org)
		} else {
			glog.Infof("Removed runner %d from %s/%s", id, c.org, c.repository)
		}
	}

	return nil
}
