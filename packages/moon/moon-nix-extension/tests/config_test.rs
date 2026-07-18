use moon_nix_extension::{EnvironmentOverrideConfig, NixExtensionConfig, OverrideModeConfig};
use moon_pdk::parse_extension_config_schema;

#[test]
fn typed_config_exposes_the_documented_defaults() {
    let config: NixExtensionConfig =
        parse_extension_config_schema(serde_json::Value::Null).unwrap();

    assert_eq!(config.base_components, ["general"]);
    assert_eq!(
        config.environment_by_toolchain.get("javascript"),
        Some(&"node".to_owned())
    );
    assert_eq!(
        config.environment_by_toolchain.get("node"),
        Some(&"node".to_owned())
    );
    assert_eq!(
        config.environment_by_toolchain.get("npm"),
        Some(&"node".to_owned())
    );
    assert_eq!(
        config.environment_by_toolchain.get("go"),
        Some(&"go".to_owned())
    );
    assert_eq!(
        config.environment_by_toolchain.get("system"),
        Some(&"general".to_owned())
    );
    assert!(config.environment_by_project.is_empty());
    assert!(config.environment_by_task.is_empty());
    assert!(config.fail_closed);
}

#[test]
fn component_override_defaults_to_append_mode() {
    let config: NixExtensionConfig = parse_extension_config_schema(serde_json::json!({
        "environmentByProject": {
            "demo": { "components": ["go"] }
        }
    }))
    .unwrap();

    let EnvironmentOverrideConfig::Components(environment) =
        config.environment_by_project.get("demo").unwrap()
    else {
        panic!("expected a component override");
    };
    assert_eq!(
        environment.components.as_ref(),
        Some(&vec!["go".to_owned()])
    );
    assert_eq!(environment.mode, OverrideModeConfig::Append);
}

#[test]
fn override_union_rejects_mixed_or_empty_shapes() {
    for environment_override in [
        serde_json::json!({}),
        serde_json::json!({ "mode": "replace" }),
        serde_json::json!({
            "components": ["go"],
            "installable": "path:./demo#go"
        }),
    ] {
        let result = parse_extension_config_schema::<NixExtensionConfig>(serde_json::json!({
            "environmentByProject": { "demo": environment_override }
        }));

        assert!(result.is_err(), "override should be rejected");
    }
}

#[test]
fn config_rejects_unknown_fields() {
    let result = parse_extension_config_schema::<NixExtensionConfig>(serde_json::json!({
        "baseComponent": ["general"]
    }));

    let message = result.unwrap_err().to_string();
    assert!(message.contains("unknown field"), "got: {message}");
}
