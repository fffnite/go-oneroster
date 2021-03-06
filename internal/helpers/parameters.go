package helpers

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"net/url"
	"strconv"
	"strings"
)

// builds the database query parameters based on user url request
// e.g. ?limit=1&fields=id
func GetOptions(q url.Values, safeFields []string) (*options.FindOptions, []error) {
	var errP []error
	projection, err := getFields(q, safeFields)
	if err != nil {
		errP = append(errP, err)
	}
	sort, err := getSort(q, safeFields)
	if err != nil {
		errP = append(errP, err)
	}
	o := options.
		Find().
		SetLimit(getLimit(q)).
		SetSkip(getOffset(q)).
		SetSort(bson.D{{sort, 1}}).
		SetProjection(projection)
	return o, errP
}

func GetOption(q url.Values, safeFields []string) (*options.FindOneOptions, []error) {
	var errP []error
	projection, err := getFields(q, safeFields)
	if err != nil {
		errP = append(errP, err)
	}
	sort, err := getSort(q, safeFields)
	if err != nil {
		errP = append(errP, err)
	}
	o := options.
		FindOne().
		SetSort(bson.D{{sort, 1}}).
		SetProjection(projection)
	return o, errP
}

// builds the filtering query based on user url request
// e.g. ?filter=id>='1'
func GetFilters(q url.Values, safeFields []string) (bson.D, error) {
	v := q.Get("filter")
	if v != "" {
		lo := parseFilterLo(v)
		var filter []bson.D
		queries := splitFilterQuery(v)
		for _, f := range queries {
			ff, err := parseFilterField(f, safeFields)
			if err != nil {
				return bson.D{}, err
			}
			fp := parseFilterPredicate(f)
			fv := parseFilterValue(f)
			doc := bson.D{{ff, bson.D{{fp, fv}}}}
			ok := checkIsoDate(fv)
			if ok {
				dfv := convertIsoDate(fv)
				doc = bson.D{{ff, bson.D{{fp, dfv}}}}
			}
			filter = append(filter, doc)
		}
		return bson.D{{lo, filter}}, nil
	}
	return bson.D{}, nil
}

// returns a bson of field filtering for mongodb from url
func getFields(q url.Values, safeFields []string) (bson.D, error) {
	v := q.Get("fields")
	d := bson.E{"_id", 0}
	var fields bson.D
	fields = append(fields, d)
	if v != "" {
		s := strings.Split(v, ",")
		for _, f := range s {
			err := validateField(f, safeFields)
			if err != nil {
				err.(*ErrorObject).CodeMinor = "invalid_selection_field"
				err.(*ErrorObject).Populate()
				return bson.D{d}, err
			}
			fields = append(fields, bson.E{f, 1})
		}
	}
	return fields, nil
}

// returns the user requested field to sort by
// validated against a field whitelist
func getSort(q url.Values, safeFields []string) (string, error) {
	v := q.Get("sort")
	d := "sourcedId"
	if v != "" {
		err := validateField(v, safeFields)
		if err != nil {
			err.(*ErrorObject).CodeMinor = "invalid_sort_field"
			err.(*ErrorObject).Populate()
			return d, err
		}
		return v, nil
	}
	return d, nil
}

// returns the max doc count requested by user
func getLimit(q url.Values) int64 {
	v := q.Get("limit")
	if v != "" {
		vi, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(err)
		}
		return vi
	}
	return 100
}

// returns the doc skip requested by user
func getOffset(q url.Values) int64 {
	v := q.Get("offset")
	if v != "" {
		vi, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			panic(err)
		}
		return vi
	}
	return 0
}
