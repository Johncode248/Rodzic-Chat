package databasepg

type Query struct {
	Limit   int     `json:"limit,omitempty"`
	OrderBy Order   `json:"order_by,omitempty"`
	Groups  []Group `json:"groups,omitempty"`
	Offset  int     `json:"offset,omitempty"`
}

func (q Query) ToSafe() Query {

	for index1 := range q.Groups {
		// Build a fresh slice rather than removing elements from
		// q.Groups[index1].Filters while ranging over it -- mutating a
		// slice mid-range (via append+re-slice) shifts later elements
		// down, so the loop's own index skips whatever moved into the
		// spot it already passed. That silently kept a "user_id" filter
		// in place when more than one was present and it wasn't the
		// first, defeating the whole point of this filter.
		kept := make([]Filter, 0, len(q.Groups[index1].Filters))
		for _, filter := range q.Groups[index1].Filters {
			if filter.Key == "user_id" {
				continue
			}
			kept = append(kept, filter)
		}
		q.Groups[index1].Filters = kept
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
