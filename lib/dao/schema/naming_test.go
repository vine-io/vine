package schema

import (
	"strings"
	"testing"
)

func TestToDBName(t *testing.T) {
	var maps = map[string]string{
		"":                          "",
		"x":                         "x",
		"X":                         "x",
		"userRestrictions":          "user_restrictions",
		"ThisIsATest":               "this_is_a_test",
		"PFAndESI":                  "pf_and_esi",
		"AbcAndJkl":                 "abc_and_jkl",
		"EmployeeID":                "employee_id",
		"SKU_ID":                    "sku_id",
		"FieldX":                    "field_x",
		"HTTPAndSMTP":               "http_and_smtp",
		"HTTPServerHandlerForURLID": "http_server_handler_for_url_id",
		"UUID":                      "uuid",
		"HTTPURL":                   "http_url",
		"HTTP_URL":                  "http_url",
		"SHA256Hash":                "sha256_hash",
		"SHA256HASH":                "sha256_hash",
		"ThisIsActuallyATestSoWeMayBeAbleToUseThisCodeInDaoPackageAlsoIdCanBeUsedAtTheEndAsID": "this_is_actually_a_test_so_we_may_be_able_to_use_this_code_in_dao_package_also_id_can_be_used_at_the_end_as_id",
	}

	ns := NamingStrategy{}
	for key, value := range maps {
		if ns.toDBName(key) != value {
			t.Errorf("%v toName should equal %v, but got %v", key, value, ns.toDBName(key))
			t.Errorf("%v", value)
			t.Errorf("%v", ns.toDBName(key))
		}
	}
}

func TestNamingStrategy(t *testing.T) {
	var ns = NamingStrategy{
		TablePrefix:   "public.",
		SingularTable: true,
		NameReplacer:  strings.NewReplacer("CID", "Cid"),
	}
	idxName := ns.IndexName("public.table", "name")

	if idxName != "idx_public_table_name" {
		t.Errorf("invalid index name generated, got %v", idxName)
	}

	chkName := ns.CheckerName("public.table", "name")
	if chkName != "chk_public_table_name" {
		t.Errorf("invalid checker name generated, got %v", chkName)
	}

	joinTable := ns.JoinTableName("user_languages")
	if joinTable != "public.user_languages" {
		t.Errorf("invalid join table generated, got %v", joinTable)
	}

	joinTable2 := ns.JoinTableName("UserLanguage")
	if joinTable2 != "public.user_language" {
		t.Errorf("invalid join table generated, got %v", joinTable2)
	}

	tableName := ns.TableName("Company")
	if tableName != "public.company" {
		t.Errorf("invalid table name generated, got %v", tableName)
	}

	columdName := ns.ColumnName("", "NameCID")
	if columdName != "name_cid" {
		t.Errorf("invalid column name generated, got %v", columdName)
	}
}

type CustomReplacer struct {
	f func(string) string
}

func (r CustomReplacer) Replace(name string) string {
	return r.f(name)
}

func TestCustomReplacer(t *testing.T) {
	var ns = NamingStrategy{
		TablePrefix:   "public.",
		SingularTable: true,
		NameReplacer: CustomReplacer{
			func(name string) string {
				replaced := "REPLACED_" + strings.ToUpper(name)
				return strings.NewReplacer("CID", "_Cid").Replace(replaced)
			},
		},
		NoLowerCase: false,
	}

	idxName := ns.IndexName("public.table", "name")
	if idxName != "idx_public_table_replaced_name" {
		t.Errorf("invalid index name generated, got %v", idxName)
	}

	chkName := ns.CheckerName("public.table", "name")
	if chkName != "chk_public_table_name" {
		t.Errorf("invalid checker name generated, got %v", chkName)
	}

	joinTable := ns.JoinTableName("user_languages")
	if joinTable != "public.user_languages" { // Seems like a bug in NamingStrategy to skip the Replacer when the name is lowercase here.
		t.Errorf("invalid join table generated, got %v", joinTable)
	}

	joinTable2 := ns.JoinTableName("UserLanguage")
	if joinTable2 != "public.replaced_userlanguage" {
		t.Errorf("invalid join table generated, got %v", joinTable2)
	}

	tableName := ns.TableName("Company")
	if tableName != "public.replaced_company" {
		t.Errorf("invalid table name generated, got %v", tableName)
	}

	columdName := ns.ColumnName("", "NameCID")
	if columdName != "replaced_name_cid" {
		t.Errorf("invalid column name generated, got %v", columdName)
	}
}

func TestCustomReplacerWithNoLowerCase(t *testing.T) {
	var ns = NamingStrategy{
		TablePrefix:   "public.",
		SingularTable: true,
		NameReplacer: CustomReplacer{
			func(name string) string {
				replaced := "REPLACED_" + strings.ToUpper(name)
				return strings.NewReplacer("CID", "_Cid").Replace(replaced)
			},
		},
		NoLowerCase: true,
	}

	idxName := ns.IndexName("public.table", "name")
	if idxName != "idx_public_table_REPLACED_NAME" {
		t.Errorf("invalid index name generated, got %v", idxName)
	}

	chkName := ns.CheckerName("public.table", "name")
	if chkName != "chk_public_table_name" {
		t.Errorf("invalid checker name generated, got %v", chkName)
	}

	joinTable := ns.JoinTableName("user_languages")
	if joinTable != "public.REPLACED_USER_LANGUAGES" {
		t.Errorf("invalid join table generated, got %v", joinTable)
	}

	joinTable2 := ns.JoinTableName("UserLanguage")
	if joinTable2 != "public.REPLACED_USERLANGUAGE" {
		t.Errorf("invalid join table generated, got %v", joinTable2)
	}

	tableName := ns.TableName("Company")
	if tableName != "public.REPLACED_COMPANY" {
		t.Errorf("invalid table name generated, got %v", tableName)
	}

	columdName := ns.ColumnName("", "NameCID")
	if columdName != "REPLACED_NAME_Cid" {
		t.Errorf("invalid column name generated, got %v", columdName)
	}
}
