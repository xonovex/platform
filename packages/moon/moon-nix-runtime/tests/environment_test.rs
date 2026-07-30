use moon_nix_runtime::environment::{
    resolve_environment, validate_environment_config, EnvironmentOverride, EnvironmentRequest,
    EnvironmentResolution, OverrideMode,
};
use std::collections::BTreeMap;

fn strings(values: &[&str]) -> Vec<String> {
    values.iter().map(|value| (*value).to_owned()).collect()
}

fn mappings(values: &[(&str, &str)]) -> BTreeMap<String, String> {
    values
        .iter()
        .map(|(toolchain, component)| ((*toolchain).to_owned(), (*component).to_owned()))
        .collect()
}

#[test]
fn base_components_alone_do_not_activate_an_unmapped_task() {
    let base_components = strings(&["general"]);
    let environment_by_toolchain = mappings(&[("node", "node")]);
    let task_toolchains = strings(&["system"]);

    let resolution = resolve_environment(EnvironmentRequest {
        base_components: &base_components,
        environment_by_toolchain: &environment_by_toolchain,
        project_override: None,
        task_override: None,
        task_toolchains: &task_toolchains,
    })
    .unwrap();

    assert_eq!(resolution, EnvironmentResolution::Unchanged);
}

#[test]
fn mapped_toolchains_activate_sorted_deduplicated_components() {
    let base_components = strings(&["general", "node"]);
    let environment_by_toolchain = mappings(&[
        ("javascript", "node"),
        ("node", "node"),
        ("npm", "node"),
        ("go", "go"),
    ]);
    let task_toolchains = strings(&["npm", "javascript", "go", "node"]);

    let resolution = resolve_environment(EnvironmentRequest {
        base_components: &base_components,
        environment_by_toolchain: &environment_by_toolchain,
        project_override: None,
        task_override: None,
        task_toolchains: &task_toolchains,
    })
    .unwrap();

    assert_eq!(
        resolution,
        EnvironmentResolution::Components(strings(&["general", "go", "node"]))
    );
}

#[test]
fn project_then_task_overrides_apply_in_specificity_order() {
    let base_components = strings(&["general"]);
    let environment_by_toolchain = mappings(&[("node", "node")]);
    let task_toolchains = strings(&["node"]);
    let project_override = EnvironmentOverride::Components {
        components: strings(&["go"]),
        mode: OverrideMode::Append,
    };
    let task_override = EnvironmentOverride::Components {
        components: strings(&["rust"]),
        mode: OverrideMode::Replace,
    };

    let resolution = resolve_environment(EnvironmentRequest {
        base_components: &base_components,
        environment_by_toolchain: &environment_by_toolchain,
        project_override: Some(&project_override),
        task_override: Some(&task_override),
        task_toolchains: &task_toolchains,
    })
    .unwrap();

    assert_eq!(
        resolution,
        EnvironmentResolution::Components(strings(&["rust"]))
    );
}

#[test]
fn scoped_override_activates_without_a_toolchain_mapping() {
    let base_components = strings(&["general"]);
    let environment_by_toolchain = BTreeMap::new();
    let task_toolchains = strings(&["system"]);
    let task_override = EnvironmentOverride::Components {
        components: strings(&["shell"]),
        mode: OverrideMode::Append,
    };

    let resolution = resolve_environment(EnvironmentRequest {
        base_components: &base_components,
        environment_by_toolchain: &environment_by_toolchain,
        project_override: None,
        task_override: Some(&task_override),
        task_toolchains: &task_toolchains,
    })
    .unwrap();

    assert_eq!(
        resolution,
        EnvironmentResolution::Components(strings(&["general", "shell"]))
    );
}

#[test]
fn empty_replace_selects_a_valid_empty_component_set() {
    let base_components = strings(&["general"]);
    let environment_by_toolchain = mappings(&[("node", "node")]);
    let task_toolchains = strings(&["node"]);
    let task_override = EnvironmentOverride::Components {
        components: vec![],
        mode: OverrideMode::Replace,
    };

    let resolution = resolve_environment(EnvironmentRequest {
        base_components: &base_components,
        environment_by_toolchain: &environment_by_toolchain,
        project_override: None,
        task_override: Some(&task_override),
        task_toolchains: &task_toolchains,
    })
    .unwrap();

    assert_eq!(resolution, EnvironmentResolution::Components(vec![]));
}

#[test]
fn task_replace_supersedes_a_project_installable() {
    let base_components = strings(&["general"]);
    let environment_by_toolchain = BTreeMap::new();
    let task_toolchains = strings(&["system"]);
    let project_override = EnvironmentOverride::Installable("path:./project#default".to_owned());
    let task_override = EnvironmentOverride::Components {
        components: strings(&["go"]),
        mode: OverrideMode::Replace,
    };

    let resolution = resolve_environment(EnvironmentRequest {
        base_components: &base_components,
        environment_by_toolchain: &environment_by_toolchain,
        project_override: Some(&project_override),
        task_override: Some(&task_override),
        task_toolchains: &task_toolchains,
    })
    .unwrap();

    assert_eq!(
        resolution,
        EnvironmentResolution::Components(strings(&["go"]))
    );
}

#[test]
fn task_append_cannot_layer_onto_a_project_installable() {
    let base_components = strings(&["general"]);
    let environment_by_toolchain = BTreeMap::new();
    let task_toolchains = strings(&["system"]);
    let project_override = EnvironmentOverride::Installable("path:./project#default".to_owned());
    let task_override = EnvironmentOverride::Components {
        components: strings(&["go"]),
        mode: OverrideMode::Append,
    };

    let error = resolve_environment(EnvironmentRequest {
        base_components: &base_components,
        environment_by_toolchain: &environment_by_toolchain,
        project_override: Some(&project_override),
        task_override: Some(&task_override),
        task_toolchains: &task_toolchains,
    })
    .unwrap_err();

    assert_eq!(
        error.to_string(),
        "task override cannot append components to a less-specific installable; use mode `replace` or select another installable"
    );
}

#[test]
fn config_validation_rejects_malformed_components() {
    let error = validate_environment_config(
        &strings(&["general", "${hostile}"]),
        &BTreeMap::new(),
        &BTreeMap::new(),
        &BTreeMap::new(),
    )
    .unwrap_err();

    assert_eq!(
        error.to_string(),
        "invalid Nix environment component `${hostile}`; expected only ASCII letters, digits, `.`, `_`, or `-`"
    );
}
