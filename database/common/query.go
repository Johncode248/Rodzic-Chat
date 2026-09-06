package databasepg

type Query struct {
	Limit   int     `json:"limit,omitempty"`
	OrderBy Order   `json:"order_by,omitempty"`
	Groups  []Group `json:"groups,omitempty"`
	Offset  int     `json:"offset,omitempty"`
}

func (q Query) ToSafe() Query {

	for index1, _ := range q.Groups {
		for index2, _ := range q.Groups[index1].Filters {
			if q.Groups[index1].Filters[index2].Key == "user_id" {
				q.Groups[index1].Filters = append(q.Groups[index1].Filters[:index2], q.Groups[index1].Filters[index2+1:]...)
			}
		}
	}

	q.Limit = 50

	return q
}

func (q Query) AddUserAuth(userId string) Query {

	if len(q.Groups) > 0 {
		for index, _ := range q.Groups {
			q.Groups[index].Filters = append(q.Groups[index].Filters, Filter{
				Key:      "user_id",
				Operator: "=",
				Value:    userId,
			})
		}
	} else {
		q.Groups = append(q.Groups, Group{
			Filters: []Filter{
				{
					Key:      "user_id",
					Operator: "=",
					Value:    userId,
				},
			},
		})
	}

	return q
}

func (q Query) AddOrganizationAuth(organizationId string) Query {

	if len(q.Groups) > 0 {
		for index, _ := range q.Groups {
			q.Groups[index].Filters = append(q.Groups[index].Filters, Filter{
				Key:      "organization_id",
				Operator: "=",
				Value:    organizationId,
			})
		}
	} else {
		q.Groups = append(q.Groups, Group{
			Filters: []Filter{
				{
					Key:      "organization_id",
					Operator: "=",
					Value:    organizationId,
				},
			},
		})
	}

	return q
}

func (q Query) AddEmailAuth(email string) Query {

	if len(q.Groups) > 0 {
		for index, _ := range q.Groups {
			q.Groups[index].Filters = append(q.Groups[index].Filters, Filter{
				Key:      "email",
				Operator: "=",
				Value:    email,
			})
		}
	} else {
		q.Groups = append(q.Groups, Group{
			Filters: []Filter{
				{
					Key:      "email",
					Operator: "=",
					Value:    email,
				},
			},
		})
	}

	return q
}
