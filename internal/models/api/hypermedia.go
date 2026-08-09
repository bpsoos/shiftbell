package api

import "encoding/json"

type Relation struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type Relations []Relation

func (relations Relations) Href(rel string) string {
	for _, relation := range relations {
		if relation.Rel == rel {
			return relation.Href
		}
	}
	return ""
}

func (relations Relations) MarshalJSON() ([]byte, error) {
	if relations == nil {
		return []byte("[]"), nil
	}
	type plainRelations Relations
	return json.Marshal(plainRelations(relations))
}
