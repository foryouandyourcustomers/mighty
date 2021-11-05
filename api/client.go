package api

import (
	"github.com/leanovate/mite-go/domain"
	"github.com/leanovate/mite-go/mite"
	log "github.com/sirupsen/logrus"
	str2duration "github.com/xhit/go-str2duration/v2"
)

type Client struct {
	api mite.Api
}

func New(baseUrl, tokenString string) (*Client, error) {

	api, err := mite.NewApi(baseUrl, tokenString, "0.0.1")
	if err != nil {
		return nil, err
	}

	return &Client{
		api,
	}, nil
}

//FetchEntries returns the past entries from the current date to given duration in the past
func (c *Client) FetchEntries(duration string) ([]*domain.TimeEntry, error) {
	dur, err := str2duration.ParseDuration(duration)
	if err != nil {
		return nil, err
	}

	to := domain.Today()
	from := to.AddDuration(-dur)

	log.Infof("Interpreting %s to fetching past entries from %s to %s ", duration, from.String(), to.String())

	entries, err := c.api.TimeEntries(&domain.TimeEntryQuery{
		UserId:    domain.CurrentUser,
		From:      &from,
		To:        &to,
		Sort:      domain.SORT_BY_PROJECT,
		Direction: domain.SORT_DIRECTION_ASC,
	})

	return entries, err
}

func (c *Client) SendEntriesToMite(entries []domain.TimeEntry) error {
	log.Infof("Pushing %d entries to mite", len(entries))

	for _, entry := range entries {

		if entry.ProjectId < 1 || entry.ServiceId < 1 {
			log.Fatalf("[%v] entry has no service id or project id, I'm ignoring it", entry)
			return nil
		}

		if entry.Id == 0 {
			log.Infof("Creating new entry [%v]", entry)

			timeEntry, err := c.api.CreateTimeEntry(&domain.TimeEntryCommand{
				Date:      &entry.Date,
				Minutes:   &entry.Minutes,
				Note:      entry.Note,
				UserId:    domain.CurrentUser,
				ProjectId: entry.ProjectId,
				ServiceId: entry.ServiceId,
				Locked:    false,
			})
			if err != nil {
				return err
			}

			log.Infof("Created new entry [%v]", timeEntry)

		} else {
			log.Infof("Editing entry [%v]", entry)

			err := c.api.EditTimeEntry(entry.Id, &domain.TimeEntryCommand{
				Date:      &entry.Date,
				Minutes:   &entry.Minutes,
				Note:      entry.Note,
				UserId:    domain.CurrentUser,
				ProjectId: entry.ProjectId,
				ServiceId: entry.ServiceId,
				Locked:    false,
			})
			if err != nil {
				return err
			}
		}

	}
	return nil
}

func (c Client) FetchServiceProjects() (map[string]domain.ServiceId, map[string]domain.ProjectId, error) {
	services, err := c.api.Services()
	if err != nil {
		return nil, nil, err
	}
	projects, err := c.api.Projects()

	if err != nil {
		return nil, nil, err
	}

	serviceIdMap := make(map[string]domain.ServiceId)

	for _, s := range services {
		serviceIdMap[s.Name] = s.Id
	}
	projectIdMap := make(map[string]domain.ProjectId)

	for _, p := range projects {
		projectIdMap[p.Name] = p.Id
	}

	return serviceIdMap, projectIdMap, nil
}
