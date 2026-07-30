use std::collections::{BTreeMap, BTreeSet};
use std::error::Error;
use std::fmt;

#[derive(Clone, Copy, Debug, Default, Eq, PartialEq)]
pub enum OverrideMode {
    #[default]
    Append,
    Replace,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum EnvironmentOverride {
    Components {
        components: Vec<String>,
        mode: OverrideMode,
    },
    Installable(String),
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum EnvironmentResolution {
    Unchanged,
    Components(Vec<String>),
    Installable(String),
}

pub struct EnvironmentRequest<'a> {
    pub base_components: &'a [String],
    pub environment_by_toolchain: &'a BTreeMap<String, String>,
    pub project_override: Option<&'a EnvironmentOverride>,
    pub task_override: Option<&'a EnvironmentOverride>,
    pub task_toolchains: &'a [String],
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct EnvironmentError(String);

impl EnvironmentError {
    fn new(message: impl Into<String>) -> Self {
        Self(message.into())
    }
}

impl fmt::Display for EnvironmentError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        formatter.write_str(&self.0)
    }
}

impl Error for EnvironmentError {}

pub fn validate_component_name(name: &str) -> Result<(), EnvironmentError> {
    let valid = !name.is_empty()
        && name
            .bytes()
            .all(|byte| byte.is_ascii_alphanumeric() || matches!(byte, b'_' | b'-' | b'.'));

    if valid {
        Ok(())
    } else {
        Err(EnvironmentError::new(format!(
            "invalid Nix environment component `{name}`; expected only ASCII letters, digits, `.`, `_`, or `-`"
        )))
    }
}

pub fn validate_environment_config(
    base_components: &[String],
    environment_by_toolchain: &BTreeMap<String, String>,
    environment_by_project: &BTreeMap<String, EnvironmentOverride>,
    environment_by_task: &BTreeMap<String, EnvironmentOverride>,
) -> Result<(), EnvironmentError> {
    validate_components(base_components)?;

    for (toolchain, component) in environment_by_toolchain {
        if toolchain.is_empty() {
            return Err(EnvironmentError::new(
                "environmentByToolchain contains an empty toolchain ID",
            ));
        }

        validate_component_name(component)?;
    }

    for environment_override in environment_by_project.values() {
        validate_override(environment_override)?;
    }

    for environment_override in environment_by_task.values() {
        validate_override(environment_override)?;
    }

    Ok(())
}

pub fn resolve_environment(
    request: EnvironmentRequest<'_>,
) -> Result<EnvironmentResolution, EnvironmentError> {
    let mapped_components = request
        .task_toolchains
        .iter()
        .filter_map(|toolchain| request.environment_by_toolchain.get(toolchain))
        .cloned()
        .collect::<Vec<_>>();
    let active = !mapped_components.is_empty()
        || request.project_override.is_some()
        || request.task_override.is_some();

    if !active {
        return Ok(EnvironmentResolution::Unchanged);
    }

    validate_components(request.base_components)?;
    validate_components(&mapped_components)?;

    let mut selected = SelectedEnvironment::Components(
        request
            .base_components
            .iter()
            .chain(&mapped_components)
            .cloned()
            .collect(),
    );

    if let Some(environment_override) = request.project_override {
        selected = apply_override(selected, environment_override, "project")?;
    }

    if let Some(environment_override) = request.task_override {
        selected = apply_override(selected, environment_override, "task")?;
    }

    Ok(match selected {
        SelectedEnvironment::Components(components) => {
            EnvironmentResolution::Components(components.into_iter().collect())
        }
        SelectedEnvironment::Installable(installable) => {
            EnvironmentResolution::Installable(installable)
        }
    })
}

enum SelectedEnvironment {
    Components(BTreeSet<String>),
    Installable(String),
}

fn validate_components(components: &[String]) -> Result<(), EnvironmentError> {
    components
        .iter()
        .try_for_each(|component| validate_component_name(component))
}

fn validate_override(environment_override: &EnvironmentOverride) -> Result<(), EnvironmentError> {
    match environment_override {
        EnvironmentOverride::Components { components, .. } => validate_components(components),
        EnvironmentOverride::Installable(installable) if installable.trim().is_empty() => Err(
            EnvironmentError::new("environment installable must not be empty"),
        ),
        EnvironmentOverride::Installable(_) => Ok(()),
    }
}

fn apply_override(
    selected: SelectedEnvironment,
    environment_override: &EnvironmentOverride,
    scope: &str,
) -> Result<SelectedEnvironment, EnvironmentError> {
    validate_override(environment_override)?;

    match environment_override {
        EnvironmentOverride::Components {
            components,
            mode: OverrideMode::Append,
        } => match selected {
            SelectedEnvironment::Components(mut selected_components) => {
                selected_components.extend(components.iter().cloned());
                Ok(SelectedEnvironment::Components(selected_components))
            }
            SelectedEnvironment::Installable(_) => Err(EnvironmentError::new(format!(
                "{scope} override cannot append components to a less-specific installable; use mode `replace` or select another installable"
            ))),
        },
        EnvironmentOverride::Components {
            components,
            mode: OverrideMode::Replace,
        } => Ok(SelectedEnvironment::Components(
            components.iter().cloned().collect(),
        )),
        EnvironmentOverride::Installable(installable) => {
            Ok(SelectedEnvironment::Installable(installable.clone()))
        }
    }
}
