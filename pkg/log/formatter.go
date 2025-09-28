package log

const (
	prefix = "fields."
)

const (
	DefaultTimestampFormat = "2006-01-02 15:04:05"

	FieldKeyMsg         = "msg"
	FieldKeyLevel       = "level"
	FieldKeyTime        = "time"
	FieldKeyLoggerError = "logger_error"
	FieldKeyEntity      = "entity"
	FieldKeyAction      = "action"
	FieldKeyMethod      = "method"
	FieldKeySubject     = "subject"
	FieldKeyData        = "data"
	FieldKeyDuration    = "duration"
	FieldKeyTraceID     = "trace_id"
	FieldKeyProtocol    = "protocol"
	FieldKeyAddress     = "addr"
	FieldKeyHeaders     = "headers"
	FieldKeyBody        = "body"
	FieldKeyRequest     = "request"
)

const (
	EntityActionIncomeRequest       = ">>"
	EntityActionOutcomeRequest      = "<<"
	EntityActionIncomeResponse      = ">"
	EntityActionOutcomeResponse     = "<"
	EntityActionTransitRequest      = ">>>"
	EntityActionTransitResponse     = "<<<"
	EntityActionOutcomeNotification = "<--"
	EntityActionIncomeNotification  = "-->"
)

type Formatter interface {
	Format(*Entry) ([]byte, error)
}

type FieldKey string

type FieldMap map[FieldKey]string

func (that FieldMap) Resolve(key FieldKey) string {
	if k, ok := that[key]; ok {
		return k
	}

	return string(key)
}

func (that FieldMap) EncodePrefixFieldClashes(data Fields) {
	that.encodePrefixFieldClash(data, FieldKeyTime)
	that.encodePrefixFieldClash(data, FieldKeyMsg)
	that.encodePrefixFieldClash(data, FieldKeyLevel)
	that.encodePrefixFieldClash(data, FieldKeyLoggerError)
}

func (that FieldMap) encodePrefixFieldClash(data Fields, key FieldKey) {
	k := that.Resolve(key)
	if l, ok := data[k]; ok {
		data[prefix+k] = l
		delete(data, k)
	}
}

func (that FieldMap) DecodePrefixFieldClashes(data Fields) {
	that.decodePrefixFieldClash(data, FieldKeyTime)
	that.decodePrefixFieldClash(data, FieldKeyMsg)
	that.decodePrefixFieldClash(data, FieldKeyLevel)
	that.decodePrefixFieldClash(data, FieldKeyLoggerError)
}

func (that FieldMap) decodePrefixFieldClash(data Fields, key FieldKey) {
	k1 := that.Resolve(key)
	k2 := prefix + k1
	if l, ok := data[k2]; ok {
		data[k1] = l
		delete(data, k2)
	}
}
