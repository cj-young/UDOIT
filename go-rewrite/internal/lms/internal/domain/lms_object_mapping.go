package domain

type LMSObjectType string

const (
	LMSObjectTypeFile LMSObjectType = "file"
)

type LMSObjectMapping struct {
	internalID int64
	objectType LMSObjectType
	lmsKey     string
	data       map[string]any
}

func NewLMSObjectMapping(internalID int64, objectType LMSObjectType, lmsKey string, data map[string]any) LMSObjectMapping {
	return LMSObjectMapping{
		internalID: internalID,
		objectType: objectType,
		lmsKey:     lmsKey,
		data:       data,
	}
}

func (m LMSObjectMapping) InternalID() int64 {
	return m.internalID
}

func (m LMSObjectMapping) ObjectType() LMSObjectType {
	return m.objectType
}

func (m LMSObjectMapping) LMSKey() string {
	return m.lmsKey
}

func (m LMSObjectMapping) Data() map[string]any {
	return m.data
}
