package api

import (
	"encoding/json"
	"fmt"
	v "github.com/gima/govalid/v1"
	"io"
	"io/ioutil"
	"log"
)

// Validate is a public function for validating APIs
func Validate(id string) {
	fmt.Printf("Import API %v\n", id)
}

func isValidJSON(input io.Reader) bool {
	var isValid bool
	schema := v.Object(
		v.ObjKV("api_model", v.Object()),
		v.ObjKV("api_definition", v.Object(
			v.ObjKV("id", v.String(v.StrRegExp("[0-9a-fA-F]+"))),
			v.ObjKV("name", v.String(v.StrMin(2))),
			v.ObjKV("api_id", v.String(v.StrRegExp("[0-9a-fA-F]+"))),
			v.ObjKV("org_id", v.String(v.StrRegExp("[0-9a-fA-F]+"))),
			v.ObjKV("use_keyless", v.Boolean()),
			v.ObjKV("use_oauth2", v.Boolean()),
			v.ObjKV("oauth_meta", v.Object(
				v.ObjKV("allowed_access_types", v.Array()),
				v.ObjKV("allowed_authorize_types", v.Array(
					v.ArrEach(v.String()),
				)),
				v.ObjKV("auth_login_redirect", v.String()),
			)),
			v.ObjKV("auth", v.Object(
				v.ObjKV("auth_header_name", v.String()),
			)),
			v.ObjKV("use_basic_auth", v.Boolean()),
			v.ObjKV("notifications", v.Object(
				v.ObjKV("shared_secret", v.String()),
				v.ObjKV("oauth_on_keychange_url", v.String()),
			)),
			v.ObjKV("enable_signature_checking", v.Boolean()),
			v.ObjKV("definition", v.Object(
				v.ObjKV("location", v.String()),
				v.ObjKV("key", v.String()),
			)),
			v.ObjKV("version_data", v.Object(
				v.ObjKV("not_versioned", v.Boolean()),
				v.ObjKV("versions", v.Object(
					v.ObjKV("Default", v.Object(
						v.ObjKV("name", v.String()),
						v.ObjKV("expires", v.String()),
						v.ObjKV("paths", v.Object(
							v.ObjKV("ignored", v.Array()),
							v.ObjKV("white_list", v.Array()),
							v.ObjKV("black_list", v.Array()),
						)),
						v.ObjKV("use_extended_paths", v.Boolean()),
						v.ObjKV("extended_paths", v.Optional(
							v.Object(
								v.ObjKV("ignored", v.Optional(v.Array(
									v.ArrEach(
										v.Object(
											v.ObjKV("path", v.String()),
											v.ObjKV("method_actions", v.Object(
												v.ObjKV(v.String(), v.Object(
													v.ObjKV("action", v.String()),
													v.ObjKV("code", v.Number()),
													v.ObjKV("data", v.String()),
													v.ObjKV("headers", v.Object()),
												)),
											)),
										),
									),
									v.ArrEach(
										v.Object(
											v.ObjKV("path", v.String()),
											v.ObjKV("method_actions", v.Object(
												v.ObjKV(v.String(), v.Object(
													v.ObjKV("action", v.String()),
													v.ObjKV("code", v.Number()),
													v.ObjKV("data", v.Object(
														v.ObjKV(v.String(), v.String()),
													)),
													v.ObjKV("headers", v.Object(
														v.ObjKV(v.String(), v.String()),
													)),
												)),
											)),
										),
									),
								))),
								v.ObjKV("white_list", v.Optional(v.Array())),
								v.ObjKV("black_list", v.Optional(v.Array())),
							),
						)),
					)),
				)),
			)),
			v.ObjKV("proxy", v.Object(
				v.ObjKV("listen_path", v.String(v.StrRegExp("^/[A-Za-z0-9-_]+/$"))),
				v.ObjKV("target_url", v.String(v.StrRegExp("^http(?s)://[A-Za-z0-9-_]+.[A-Za-z0-9-_]+?(/)$"))),
				v.ObjKV("strip_listen_path", v.Boolean()),
			)),
			v.ObjKV("custom_middleware", v.Object(
				v.ObjKV("pre", v.Array()),
				v.ObjKV("post", v.Array()),
			)),
			v.ObjKV("session_lifetime", v.Number()),
			v.ObjKV("active", v.Boolean()),
			v.ObjKV("auth_provider", v.Object(
				v.ObjKV("name", v.Optional(v.String())),
				v.ObjKV("storage_engine", v.Optional(v.String())),
				v.ObjKV("meta", v.Or(v.Nil(), v.Object())),
			)),
			v.ObjKV("session_provider", v.Object(
				v.ObjKV("name", v.Optional(v.String())),
				v.ObjKV("storage_engine", v.Optional(v.String())),
				v.ObjKV("meta", v.Or(v.Nil(), v.Object())),
			)),
			v.ObjKV("event_handlers", v.Object(
				v.ObjKV("events", v.Optional(
					v.Object(
						v.ObjKV("QuotaExceeded", v.Optional(
							v.Array(
								v.ArrEach(
									v.Object(
										v.ObjKV("handler_name", v.String()),
										v.ObjKV("handler_meta", v.Object(
											v.ObjKV("_id", v.String(v.StrRegExp("[0-9a-fA-F]+"))),
											v.ObjKV("event_timeout", v.Number()),
											v.ObjKV("header_map", v.Object(
												v.ObjKV(
													v.String(), v.String(),
												),
											)),
											v.ObjKV("method", v.String(v.StrRegExp("[A-Z]+"))),
											v.ObjKV("name", v.String()),
											v.ObjKV("org_id", v.String(v.StrRegExp("[0-9a-fA-F]+"))),
											v.ObjKV("target_path", v.String(v.StrRegExp("^http(?s)://w+.[a-zA-Z]+(?.[a-zA-Z]+)/w+$"))),
											v.ObjKV("template_path", v.Optional(v.String())),
										)),
									),
								),
							),
						)),
					),
				)),
			)),
			v.ObjKV("enable_batch_request_support", v.Boolean()),
			v.ObjKV("enable_ip_whitelisting", v.Boolean()),
			v.ObjKV("allowed_ips", v.Array(
				v.ArrEach(
					v.String(v.StrRegExp("^[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3}$")),
				),
			)),
			v.ObjKV("expire_analytics_after", v.Number()),
		)),
		v.ObjKV("hook_references", v.Array(
			v.ArrEach(
				v.Object(
					v.ObjKV("event_name", v.String()),
					v.ObjKV("event_timeout", v.Number()),
					v.ObjKV("hook", v.Object(
						v.ObjKV("api_model", v.Object()),
						v.ObjKV("id", v.String(v.StrRegExp("[0-9a-fA-F]+"))),
						v.ObjKV("org_id", v.String(v.StrRegExp("[0-9a-fA-F]+"))),
						v.ObjKV("name", v.String()),
						v.ObjKV("method", v.String(v.StrRegExp("[A-Z]+"))),
						v.ObjKV("target_path", v.String(v.StrRegExp("^http(?s)://w+.[a-zA-Z]+(?.[a-zA-Z]+)/w+$"))),
						v.ObjKV("template_path", v.String()),
						v.ObjKV("header_map", v.Object(
							v.ObjKV(v.String(), v.String()),
						)),
						v.ObjKV("event_timeout", v.Number()),
					)),
				),
			),
		)),
	)
	var inputJSON interface{}
	inputByte, _ := ioutil.ReadAll(input)
	err := json.Unmarshal(inputByte, &inputJSON)
	if err != nil {
		log.Printf("Error: %v", err)
		return false
	}
	if path, err := schema.Validate(inputJSON.(map[string]interface{})); err == nil {
		log.Print("Validation passed.")
		isValid = true
	} else {
		log.Printf("Validation failed at %s. Error (%s)", path, err)
		isValid = false
	}
	return isValid
}
