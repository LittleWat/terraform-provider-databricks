package common

import (
	"reflect"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

type DummyTfSdk struct {
	Enabled           types.Bool                  `tfsdk:"enabled" tf:"optional"`
	Workers           types.Int64                 `tfsdk:"workers" tf:""` // Test required field
	Floats            types.Float64               `tfsdk:"floats" tf:""`  // Test required field
	Description       types.String                `tfsdk:"description" tf:""`
	Tasks             types.String                `tfsdk:"task" tf:"optional"`
	Nested            *DummyNestedTfSdk           `tfsdk:"nested" tf:"optional"`
	NoPointerNested   DummyNestedTfSdk            `tfsdk:"no_pointer_nested" tf:"optional"`
	NestedList        []DummyNestedTfSdk          `tfsdk:"nested_list" tf:"optional"`
	NestedPointerList []*DummyNestedTfSdk         `tfsdk:"nested_pointer_list" tf:"optional"`
	Map               map[string]types.String     `tfsdk:"map" tf:"optional"`
	NestedMap         map[string]DummyNestedTfSdk `tfsdk:"nested_map" tf:"optional"`
	Repeated          []types.Int64               `tfsdk:"repeated" tf:"optional"`
	Attributes        map[string]types.String     `tfsdk:"attributes" tf:"optional"`
	EnumField         TestEnum                    `tfsdk:"enum_field" tf:"optional"`
	Irrelevant        types.String                `tfsdk:"-"`
}

type TestEnum string

const TestEnumA TestEnum = `TEST_ENUM_A`

const TestEnumB TestEnum = `TEST_ENUM_B`

const TestEnumC TestEnum = `TEST_ENUM_C`

type DummyNestedTfSdk struct {
	Name    types.String `tfsdk:"name" tf:"optional"`
	Enabled types.Bool   `tfsdk:"enabled" tf:"optional"`
}

type DummyGoSdk struct {
	Enabled           bool                        `json:"enabled"`
	Workers           int64                       `json:"workers"`
	Floats            float64                     `json:"floats"`
	Description       string                      `json:"description"`
	Tasks             string                      `json:"tasks"`
	Nested            *DummyNestedGoSdk           `json:"nested"`
	NoPointerNested   DummyNestedGoSdk            `json:"no_pointer_nested"`
	NestedList        []DummyNestedGoSdk          `json:"nested_list"`
	NestedPointerList []*DummyNestedGoSdk         `json:"nested_pointer_list"`
	Map               map[string]string           `json:"map"`
	NestedMap         map[string]DummyNestedGoSdk `json:"nested_map"`
	Repeated          []int64                     `json:"repeated"`
	Attributes        map[string]string           `json:"attributes"`
	EnumField         TestEnum                    `json:"enum_field"`
	ForceSendFields   []string                    `json:"-"`
}

type DummyNestedGoSdk struct {
	Name            string   `json:"name"`
	Enabled         bool     `json:"enabled"`
	ForceSendFields []string `json:"-"`
}

type emptyCtx struct{}

func (emptyCtx) Deadline() (deadline time.Time, ok bool) {
	return
}

func (emptyCtx) Done() <-chan struct{} {
	return nil
}

func (emptyCtx) Err() error {
	return nil
}

func (emptyCtx) Value(key any) any {
	return nil
}

var ctx = emptyCtx{}

// Constructing a dummy tfsdk struct.
var tfSdkStruct = DummyTfSdk{
	Enabled:     types.BoolValue(false),
	Workers:     types.Int64Value(12),
	Description: types.StringValue("abc"),
	Tasks:       types.StringNull(),
	Nested: &DummyNestedTfSdk{
		Name:    types.StringValue("def"),
		Enabled: types.BoolValue(true),
	},
	NoPointerNested: DummyNestedTfSdk{
		Name:    types.StringValue("def"),
		Enabled: types.BoolValue(true),
	},
	NestedList: []DummyNestedTfSdk{
		{
			Name:    types.StringValue("def"),
			Enabled: types.BoolValue(true),
		},
		{
			Name:    types.StringValue("def"),
			Enabled: types.BoolValue(true),
		},
	},
	Map: map[string]types.String{
		"key1": types.StringValue("value1"),
		"key2": types.StringValue("value2"),
	},
	NestedMap: map[string]DummyNestedTfSdk{
		"key1": {
			Name:    types.StringValue("abc"),
			Enabled: types.BoolValue(false),
		},
		"key2": {
			Name:    types.StringValue("def"),
			Enabled: types.BoolValue(true),
		},
	},
	NestedPointerList: []*DummyNestedTfSdk{
		{
			Name:    types.StringValue("def"),
			Enabled: types.BoolValue(true),
		},
		{
			Name:    types.StringValue("def"),
			Enabled: types.BoolValue(true),
		},
	},
	Attributes: map[string]types.String{"key": types.StringValue("value")},
	EnumField:  TestEnumA,
	Repeated:   []types.Int64{types.Int64Value(12), types.Int64Value(34)},
}

func TestGetAndSetPluginFramework(t *testing.T) {
	// Also test StructToSchema.
	scm := PluginFrameworkResourceStructToSchema(DummyTfSdk{}, nil)
	state := tfsdk.State{
		Schema: scm,
	}

	// Test optional is being handled properly.
	assert.True(t, scm.Attributes["enabled"].IsOptional())
	assert.True(t, scm.Attributes["workers"].IsRequired())
	assert.True(t, scm.Attributes["floats"].IsRequired())

	// Assert that we can set state from the tfsdk struct.
	diags := state.Set(ctx, tfSdkStruct)
	assert.Len(t, diags, 0)

	// Assert that we can get a struct from the state.
	getterStruct := DummyTfSdk{}
	diags = state.Get(ctx, &getterStruct)
	assert.Len(t, diags, 0)

	// Assert the struct populated from .Get is exactly the same as the original tfsdk struct.
	assert.True(t, reflect.DeepEqual(getterStruct, tfSdkStruct))
}

func TestStructConversion(t *testing.T) {
	// Convert from tfsdk to gosdk struct using the converter function
	convertedGoSdkStruct := DummyGoSdk{}
	e := TfSdkToGoSdkStruct(tfSdkStruct, &convertedGoSdkStruct, ctx)
	assert.NoError(t, e)

	// Convert the gosdk struct back to tfsdk struct
	convertedTfSdkStruct := DummyTfSdk{}
	e = GoSdkToTfSdkStruct(convertedGoSdkStruct, &convertedTfSdkStruct, ctx)
	assert.NoError(t, e)

	// Assert that the struct is exactly the same after tfsdk --> gosdk --> tfsdk
	assert.True(t, reflect.DeepEqual(tfSdkStruct, convertedTfSdkStruct))
}