package transformer

import (
	"encoding/json"
	"strings"

	"github.com/vietquy/alpha/errors"
	"github.com/vietquy/alpha/messaging"
)

const sep = "/"

var keys = [...]string{"publisher", "protocol", "project", "subtopic"}

var (
	// ErrTransform reprents an error during parsing message.
	ErrTransform         = errors.New("unable to parse JSON object")
	errInvalidKey        = errors.New("invalid object key")
	errUnknownFormat     = errors.New("unknown format of JSON message")
	errInvalidFormat     = errors.New("invalid JSON object")
	errInvalidNestedJSON = errors.New("invalid nested JSON object")
)

// Transformer specifies API form Message transformer.
type Transformer interface {
	// Transform message to any other format.
	Transform(msg messaging.Message) (interface{}, error)
}

type typeTransformer func(messaging.Message) (interface{}, error)

// New returns a new JSON transformer.
func New() Transformer {
	return typeTransformer(funcTransformer)
}

func (fh typeTransformer) Transform(msg messaging.Message) (interface{}, error) {
	return fh(msg)
}

func funcTransformer(msg messaging.Message) (interface{}, error) {
	ret := Message{
		Publisher: msg.Publisher,
		Created:   msg.Created,
		Protocol:  msg.Protocol,
		Project:   msg.Project,
		Subtopic:  msg.Subtopic,
	}
	subs := strings.Split(ret.Subtopic, ".")
	if len(subs) == 0 {
		return nil, errors.Wrap(ErrTransform, errUnknownFormat)
	}
	format := subs[len(subs)-1]
	var payload interface{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return nil, errors.Wrap(ErrTransform, err)
	}
	switch p := payload.(type) {
	case map[string]interface{}:
		flat, err := Flatten(p)
		if err != nil {
			return nil, errors.Wrap(ErrTransform, err)
		}
		ret.Payload = flat
		return Messages{[]Message{ret}, format}, nil
	case []interface{}:
		res := []Message{}
		// Make an array of messages from the root array.
		for _, val := range p {
			v, ok := val.(map[string]interface{})
			if !ok {
				return nil, errors.Wrap(ErrTransform, errInvalidNestedJSON)
			}
			flat, err := Flatten(v)
			if err != nil {
				return nil, errors.Wrap(ErrTransform, err)
			}
			newMsg := ret
			newMsg.Payload = flat
			res = append(res, newMsg)
		}
		return Messages{res, format}, nil
	default:
		return nil, errors.Wrap(ErrTransform, errInvalidFormat)
	}
}

// ParseFlat receives flat map that reprents complex JSON objects and returns
// the corresponding complex JSON object with nested maps. It's the opposite
// of the Flatten function.
func ParseFlat(flat interface{}) interface{} {
	msg := make(map[string]interface{})
	switch v := flat.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if value == nil {
				continue
			}
			keys := strings.Split(key, sep)
			n := len(keys)
			if n == 1 {
				msg[key] = value
				continue
			}
			current := msg
			for i, k := range keys {
				if _, ok := current[k]; !ok {
					current[k] = make(map[string]interface{})
				}
				if i == n-1 {
					current[k] = value
					break
				}
				current = current[k].(map[string]interface{})
			}
		}
	}
	return msg
}

// Flatten makes nested maps flat using composite keys created by concatenation of the nested keys.
func Flatten(m map[string]interface{}) (map[string]interface{}, error) {
	return flatten("", make(map[string]interface{}), m)
}

func flatten(prefix string, m, m1 map[string]interface{}) (map[string]interface{}, error) {
	for k, v := range m1 {
		if strings.Contains(k, sep) {
			return nil, errInvalidKey
		}
		for _, key := range keys {
			if k == key {
				return nil, errInvalidKey
			}
		}
		switch val := v.(type) {
		case map[string]interface{}:
			var err error
			m, err = flatten(prefix+k+sep, m, val)
			if err != nil {
				return nil, err
			}
		default:
			m[prefix+k] = v
		}
	}
	return m, nil
}
