package influxdb

import (
	// "math"
	"time"

	"github.com/vietquy/alpha/writer"
	"github.com/vietquy/alpha/errors"
	"github.com/vietquy/alpha/transformer"

	influxdata "github.com/influxdata/influxdb/client/v2"
)

var errSaveMessage = errors.New("failed to save message to influxdb database")

var _ writer.Writer = (*influxRepo)(nil)

type tags map[string]string

type influxRepo struct {
	client influxdata.Client
	cfg    influxdata.BatchPointsConfig
}

// New returns new InfluxDB writer.
func New(client influxdata.Client, database string) writer.Writer {
	return &influxRepo{
		client: client,
		cfg: influxdata.BatchPointsConfig{
			Database: database,
		},
	}
}

func (repo *influxRepo) Write(message interface{}) error {
	pts, err := influxdata.NewBatchPoints(repo.cfg)
	if err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	pts, err = repo.jsonPoints(pts, message.(transformer.Messages))
	if err != nil {
		return err
	}

	if err := repo.client.Write(pts); err != nil {
		return errors.Wrap(errSaveMessage, err)
	}
	return nil
}

func (repo *influxRepo) jsonPoints(pts influxdata.BatchPoints, msgs transformer.Messages) (influxdata.BatchPoints, error) {
	for i, m := range msgs.Data {
		t := time.Unix(0, m.Created+int64(i))

		// Copy first-level fields so that the original Payload is unchanged.
		fields := make(map[string]interface{})
		for k, v := range m.Payload {
			fields[k] = v
		}
		// At least one known field need to exist so that COUNT can be performed.
		fields["protocol"] = m.Protocol
		pt, err := influxdata.NewPoint(msgs.Format, jsonTags(m), fields, t)
		if err != nil {
			return nil, errors.Wrap(errSaveMessage, err)
		}
		pts.AddPoint(pt)
	}

	return pts, nil
}

func jsonTags(msg transformer.Message) tags {
	return tags{
		"project":   msg.Project,
		"subtopic":  msg.Subtopic,
		"publisher": msg.Publisher,
	}
}
