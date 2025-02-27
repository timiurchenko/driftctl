package filter

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/snyk/driftctl/pkg/resource"
)

func TestDriftIgnore_IsResourceIgnored(t *testing.T) {
	tests := []struct {
		name      string
		resources []*resource.Resource
		want      []bool
		path      string
		ignores   []string
	}{
		{
			name: "drift_ignore_no_file",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
			},
			want: []bool{
				false,
			},
			path: "testdata/drift_ignore_no_file/.driftignore",
		},
		{
			name: "drift_ignore_empty",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
			},
			want: []bool{
				false,
			},
			path: "testdata/drift_ignore_empty/.driftignore",
		},
		{
			name: "drift_ignore_invalid_lines",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "ignored_resource",
					Id:   "id2",
				},
			},
			want: []bool{
				false,
				true,
			},
			path: "testdata/drift_ignore_invalid_lines/.driftignore",
		},
		{
			name: "drift_ignore_valid",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "wildcard_resource",
					Id:   "id1/with/slash",
				},
				{
					Type: "wildcard_resource",
					Id:   "id1",
				},
				{
					Type: "wildcard_resource",
					Id:   "id2",
				},
				{
					Type: "wildcard_resource",
					Id:   "id3",
				},
				{
					Type: "ignored_resource",
					Id:   "id2",
				},
				{
					Type: "resource_type",
					Id:   "id.with.dots",
				},
				{
					Type: "resource_type",
					Id:   "idwith\\",
				},
				{
					Type: "resource_type",
					Id:   "idwith\\backslashes",
				},
				{
					Type: "resource_type",
					Id:   "idwith/slashes",
				},
			},
			want: []bool{
				false,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
				true,
			},
			path: "testdata/drift_ignore_valid/.driftignore",
		},
		{
			name: "drift_ignore_wildcard",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id11",
				},
				{
					Type: "type2",
					Id:   "id2",
				},
				{
					Type: "type3",
					Id:   "id100",
				},
				{
					Type: "type3",
					Id:   "id101",
				},
				{
					Type: "type4",
					Id:   "id\\WithBac*slash***\\*\\",
				},
			},
			want: []bool{
				false,
				true,
				true,
				false,
				true,
				false,
				true,
			},
			path: "testdata/drift_ignore_wildcard/.driftignore",
		},
		{
			name: "drift_ignore_all_exclude",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id11",
				},
				{
					Type: "type2",
					Id:   "id2",
				},
				{
					Type: "type3",
					Id:   "id100",
				},
				{
					Type: "type3",
					Id:   "id101",
				},
				{
					Type: "iam_user",
					Id:   "id\\WithBac*slash***\\*\\",
				},
				{
					Type: "some_type",
					Id:   "idwith/slash",
				},
				{
					Type: "some_type",
					Id:   "idwith/slash/",
				},
			},
			want: []bool{
				true,
				true,
				true,
				true,
				true,
				true,
				false,
				false,
				true,
			},
			path: "testdata/drift_ignore_all_exclude/.driftignore",
		},
		{
			name: "drift_ignore_all_exclude_with_ignore_patterns",
			resources: []*resource.Resource{
				{
					Type: "type1",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id1",
				},
				{
					Type: "type2",
					Id:   "id11",
				},
				{
					Type: "type2",
					Id:   "id2",
				},
				{
					Type: "type3",
					Id:   "id100",
				},
				{
					Type: "type3",
					Id:   "id101",
				},
				{
					Type: "iam_user",
					Id:   "id\\WithBac*slash***\\*\\",
				},
				{
					Type: "some_type",
					Id:   "idwith/slash",
				},
				{
					Type: "some_type",
					Id:   "idwith/slash/",
				},
			},
			want: []bool{
				true,
				true,
				true,
				true,
				true,
				true,
				false,
				false,
				true,
			},
			path:    "testdata/drift_ignore_all/.driftignore",
			ignores: []string{"*", "!iam_user.*", "!some_type.idwith/slash"},
		},
		{
			name: "drift_ignore_none_with_ignore_patterns",
			resources: []*resource.Resource{
				{
					Type: "aws_s3_access_point",
				},
			},
			want: []bool{
				false,
			},
			path:    "testdata/drift_ignore_all/.driftignore",
			ignores: []string{"!*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwd, _ := os.Getwd()
			defer func() { _ = os.Chdir(cwd) }()

			r := NewDriftIgnore(tt.path, tt.ignores...)
			got := make([]bool, 0, len(tt.want))
			for _, res := range tt.resources {
				got = append(got, r.IsResourceIgnored(res))
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDriftIgnore_IsFieldIgnored(t *testing.T) {

	type Args struct {
		Res  *resource.Resource
		Path []string
		Want bool
	}

	tests := []struct {
		name    string
		args    []Args
		path    string
		ignores []string
	}{
		{
			name: "drift_ignore_no_file",
			args: []Args{

				{
					Res:  &resource.Resource{Type: "type1", Id: "id1"},
					Path: []string{"Id"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "type2", Id: "id2"},
					Path: []string{"Id"},
					Want: false,
				},
			},
			path: "testdata/drift_ignore_no_file/.driftignore",
		},
		{
			name: "drift_ignore_empty",
			args: []Args{
				{
					Res:  &resource.Resource{Type: "type1", Id: "id1"},
					Path: []string{"Id"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "type2", Id: "id2"},
					Path: []string{"Id"},
					Want: false,
				},
			},
			path: "testdata/drift_ignore_empty/.driftignore",
		},
		{
			name: "drift_ignore_fields",
			args: []Args{
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"json"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"json"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: true,
				},
			},
			path: "testdata/drift_ignore_fields/.driftignore",
		},
		{
			name: "drift_ignore_all_exclude_field",
			args: []Args{
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
			},
			path: "testdata/drift_ignore_all_exclude_field/.driftignore",
		},
		{
			name: "drift_ignore_all_exclude_field_with_ignore_patterns",
			args: []Args{
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "full_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "partial_drift_ignored"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "id.with.dots"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"json"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "resource_type", Id: "idwith\\backslashes"},
					Path: []string{"foobar"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "wildcard_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "baz"},
					Want: true,
				},
				{
					Res:  &resource.Resource{Type: "res_type", Id: "endofpath_drift_ignored"},
					Path: []string{"struct", "bar"},
					Want: false,
				},
			},
			path:    "testdata/drift_ignore_all/.driftignore",
			ignores: []string{"*", "!*.bar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwd, _ := os.Getwd()
			defer func() { _ = os.Chdir(cwd) }()

			r := NewDriftIgnore(tt.path, tt.ignores...)
			for _, arg := range tt.args {
				got := r.IsFieldIgnored(arg.Res, arg.Path)
				if arg.Want != got {
					t.Errorf("%s.%s.%s expected %v got %v", arg.Res.ResourceType(), arg.Res.ResourceId(), strings.Join(arg.Path, "."), arg.Want, got)
				}
			}
		})
	}
}

func TestDriftIgnore_IsTypeIgnored(t *testing.T) {
	tests := []struct {
		name      string
		resources []*resource.Resource
		want      []bool
		path      string
		ignores   []string
	}{
		{
			name: "drift_ignore_type_exclude_with_child_1_nesting",
			resources: []*resource.Resource{
				{
					Type: "aws_route",
				},
				{
					Type: "aws_route_table",
				},
				{
					Type: "non_ignored_type",
				},
				{
					Type: "ignored_type",
				},
			},
			want: []bool{
				false,
				false,
				false,
				true,
			},
			path: "testdata/drift_ignore_type/.driftignore_child_1",
		},
		{
			name: "drift_ignore_type_exclude_with_child_2_nesting",
			resources: []*resource.Resource{
				{
					Type: "non_ignored_type",
				},
				{
					Type: "aws_iam_user",
				},
				{
					Type: "aws_iam_user_policy",
				},
				{
					Type: "aws_iam_user_policy_attachment",
				},
				{
					Type: "ignored_type",
				},
			},
			want: []bool{
				false,
				false,
				false,
				false,
				true,
			},
			path: "testdata/drift_ignore_type/.driftignore_child_2",
		},
		{
			name: "drift_ignore_type_exclude",
			resources: []*resource.Resource{
				{
					Type: "type",
				},
				{
					Type: "type_1",
				},
				{
					Type: "type_2",
				},
				{
					Type: "type_3",
				},
			},
			want: []bool{
				true,
				false,
				true,
				true,
			},
			path: "testdata/drift_ignore_type/.driftignore",
		},
		{
			name: "drift_ignore_non_aws_s3_resources",
			resources: []*resource.Resource{
				{
					Type: "aws_s3_access_point",
				},
				{
					Type: "aws_s3_bucket",
				},
				{
					Type: "aws_s3_bucket_acl",
				},
				{
					Type: "aws_route53_delegation_set",
				},
			},
			want: []bool{
				false,
				false,
				false,
				true,
			},
			path:    "testdata/drift_ignore_all/.driftignore",
			ignores: []string{"*", "!aws_s3*"},
		},
		{
			name: "drift_ignore_non_aws_s3_and_non_route53_resources",
			resources: []*resource.Resource{
				{
					Type: "aws_s3_access_point",
				},
				{
					Type: "aws_s3_bucket",
				},
				{
					Type: "aws_s3_bucket_acl",
				},
				{
					Type: "aws_route53_delegation_set",
				},
			},
			want: []bool{
				false,
				false,
				false,
				false,
			},
			path:    "testdata/drift_ignore_all/.driftignore",
			ignores: []string{"*", "!aws_s3*", "!aws_route53*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cwd, _ := os.Getwd()
			defer func() { _ = os.Chdir(cwd) }()

			r := NewDriftIgnore(tt.path, tt.ignores...)
			got := make([]bool, 0, len(tt.want))
			for _, res := range tt.resources {
				got = append(got, r.IsTypeIgnored(resource.ResourceType(res.ResourceType())))
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
