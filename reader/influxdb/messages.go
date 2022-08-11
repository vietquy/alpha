package influxdb

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/vietquy/alpha/errors"
	"github.com/vietquy/alpha/reader"

	influxdata "github.com/influxdata/influxdb/client/v2"
	"github.com/vietquy/alpha/transformer"
)

const (
	countCol = "count_protocol"
	defMeasurement = "messages"
)

var errReadMessages = errors.New("failed to read messages from influxdb database")

var _ reader.MessageRepository = (*influxRepository)(nil)

type influxRepository struct {
	database string
	client   influxdata.Client
}

// New returns new InfluxDB reader.
func New(client influxdata.Client, database string) reader.MessageRepository {
	return &influxRepository{
		database,
		client,
	}
}

func (repo *influxRepository) ReadAll(projectID string, rpm reader.PageMetadata) (reader.MessagesPage, error) {
	format := defMeasurement
	if rpm.Format != "" {
		format = rpm.Format
	}

	condition := fmtCondition(projectID, rpm)

	cmd := fmt.Sprintf(`SELECT * FROM %s WHERE %s ORDER BY time DESC LIMIT %d OFFSET %d`, format, condition, rpm.Limit, rpm.Offset)
	q := influxdata.Query{
		Command:  cmd,
		Database: repo.database,
	}

	var ret []reader.Message

	resp, err := repo.client.Query(q)
	if err != nil {
		return reader.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}
	if resp.Error() != nil {
		return reader.MessagesPage{}, errors.Wrap(errReadMessages, resp.Error())
	}

	if len(resp.Results) < 1 || len(resp.Results[0].Series) < 1 {
		return reader.MessagesPage{}, nil
	}

	result := resp.Results[0].Series[0]
	for _, v := range result.Values {
		ret = append(ret, parseJSON(result.Columns, v))
	}

	total, err := repo.count(format, condition)
	if err != nil {
		return reader.MessagesPage{}, errors.Wrap(errReadMessages, err)
	}

	page := reader.MessagesPage{
		PageMetadata: rpm,
		Total:        total,
		Messages:     ret,
	}

	return page, nil
}

func (repo *influxRepository) count(measurement, condition string) (uint64, error) {
	cmd := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE %s`, measurement, condition)
	q := influxdata.Query{
		Command:  cmd,
		Database: repo.database,
	}

	resp, err := repo.client.Query(q)
	if err != nil {
		return 0, err
	}
	if resp.Error() != nil {
		return 0, resp.Error()
	}

	if len(resp.Results) < 1 ||
		len(resp.Results[0].Series) < 1 ||
		len(resp.Results[0].Series[0].Values) < 1 {
		return 0, nil
	}

	countIndex := 0
	for i, col := range resp.Results[0].Series[0].Columns {
		if col == countCol {
			countIndex = i
			break
		}
	}

	result := resp.Results[0].Series[0].Values[0]
	if len(result) < countIndex+1 {
		return 0, nil
	}

	count, ok := result[countIndex].(json.Number)
	if !ok {
		return 0, nil
	}

	return strconv.ParseUint(count.String(), 10, 64)
}

func fmtCondition(projectID string, rpm reader.PageMetadata) string {
	condition := fmt.Sprintf(`project='%s'`, projectID)

	var query map[string]interface{}
	meta, err := json.Marshal(rpm)
	if err != nil {
		return condition
	}
	json.Unmarshal(meta, &query)

	for name, value := range query {
		switch name {
		case
			"project",
			"subtopic",
			"publisher",
			"name",
			"protocol":
			condition = fmt.Sprintf(`%s AND "%s"='%s'`, condition, name, value)
		case "v":
			comparator := reader.ParseValueComparator(query)
			condition = fmt.Sprintf(`%s AND value %s %f`, condition, comparator, value)
		case "vb":
			condition = fmt.Sprintf(`%s AND boolValue = %t`, condition, value)
		case "vs":
			condition = fmt.Sprintf(`%s AND stringValue = '%s'`, condition, value)
		case "vd":
			condition = fmt.Sprintf(`%s AND dataValue = '%s'`, condition, value)
		case "from":
			iVal := int64(value.(float64) * 1e9)
			condition = fmt.Sprintf(`%s AND time >= %d`, condition, iVal)
		case "to":
			iVal := int64(value.(float64) * 1e9)
			condition = fmt.Sprintf(`%s AND time < %d`, condition, iVal)
		}
	}
	return condition
}


func parseJSON(names []string, fields []interface{}) interface{} {
	ret := make(map[string]interface{})
	for i, n := range names {
		ret[n] = fields[i]
	}

	return transformer.ParseFlat(ret)
}
