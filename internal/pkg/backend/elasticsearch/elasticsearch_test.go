package elasticsearch

import (
	"testing"
)

func TestElasticSearch_HasSent(t *testing.T) {
	t.Parallel()

	type fields struct {
		Documents       []ElasticSearchDocument
		SentDocumentIDs []string
	}

	type args struct {
		document *ElasticSearchDocument
	}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "ensure missing document returns false",
			fields: fields{
				SentDocumentIDs: []string{"1", "2", "3", "5"},
			},
			args: args{
				&ElasticSearchDocument{
					id: "4",
				},
			},
			want: false,
		},
		{
			name: "ensure existing document returns true",
			fields: fields{
				SentDocumentIDs: []string{"1", "2", "3", "4", "5"},
			},
			args: args{
				&ElasticSearchDocument{
					id: "3",
				},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			es := &ElasticSearch{
				SentDocumentIDs: tt.fields.SentDocumentIDs,
			}
			if got := es.HasSent(tt.args.document); got != tt.want {
				t.Errorf("ElasticSearch.HasSent() = %v, want %v", got, tt.want)
			}
		})
	}
}
