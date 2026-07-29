package domain

type LMSObjectType string

const (
	LMSObjectTypeFile LMSObjectType = "file"
)


// `externalID` oes not necessarily represent the ID of the object in the LMS.
// It is simply a unique identifier for some resource within a tenant, the form
// being determined by a provider. 
type LMSObjectMapping struct {
	internalID int64
	objectType LMSObjectType
	externalID string 
	lmsKey     string
	data       map[string]any
}

func NewLMSObjectMapping(internalID int64, objectType LMSObjectType, externalID string, lmsKey string, data map[string]any) LMSObjectMapping {
	return LMSObjectMapping{
		internalID: internalID,
		objectType: objectType,
		externalID: externalID,
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

func (m LMSObjectMapping) ExternalID() string {
	return m.externalID
}

func (m LMSObjectMapping) Data() map[string]any {
	return m.data
}
