use crate::environment::{validate_component_name, EnvironmentError};
use std::collections::BTreeMap;
use std::error::Error;
use std::fmt;
use std::path::{Path, PathBuf};

pub const MOON_FLAKE_ENV: &str = "MOON_NIX_FLAKE";
pub const MOON_COMPONENTS_ENV: &str = "MOON_NIX_COMPONENTS_JSON";
pub const MOON_SHELL_EXPRESSION: &str = r#"let flake = builtins.getFlake (builtins.getEnv "MOON_NIX_FLAKE"); components = builtins.fromJSON (builtins.getEnv "MOON_NIX_COMPONENTS_JSON"); in { moon = flake.lib.mkMoonShell builtins.currentSystem components; }"#;

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct ParsedInstallable {
    pub relative_flake: PathBuf,
    pub dev_shell: String,
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub struct InstallableError(String);

impl InstallableError {
    fn new(message: impl Into<String>) -> Self {
        Self(message.into())
    }
}

impl fmt::Display for InstallableError {
    fn fmt(&self, formatter: &mut fmt::Formatter<'_>) -> fmt::Result {
        formatter.write_str(&self.0)
    }
}

impl Error for InstallableError {}

impl From<EnvironmentError> for InstallableError {
    fn from(error: EnvironmentError) -> Self {
        Self::new(error.to_string())
    }
}

pub fn parse_workspace_installable(value: &str) -> Result<ParsedInstallable, InstallableError> {
    let Some(path_and_shell) = value.strip_prefix("path:") else {
        return Err(InstallableError::new(format!(
            "invalid environment installable `{value}`; expected `path:./relative#devShell`"
        )));
    };
    let Some((relative_flake, dev_shell)) = path_and_shell.split_once('#') else {
        return Err(InstallableError::new(format!(
            "invalid environment installable `{value}`; an explicit devShell after `#` is required"
        )));
    };

    if relative_flake.is_empty()
        || !(relative_flake == "." || relative_flake.starts_with("./"))
        || relative_flake.contains('#')
    {
        return Err(InstallableError::new(format!(
            "invalid environment installable `{value}`; the flake must be workspace-relative (`path:./...`)"
        )));
    }

    if dev_shell.is_empty() || dev_shell.contains('#') {
        return Err(InstallableError::new(format!(
            "invalid environment installable `{value}`; exactly one explicit devShell is required"
        )));
    }

    validate_component_name(dev_shell)?;

    Ok(ParsedInstallable {
        relative_flake: PathBuf::from(relative_flake),
        dev_shell: dev_shell.to_owned(),
    })
}

pub fn canonical_installable(
    canonical_workspace: &Path,
    canonical_flake: &Path,
    dev_shell: &str,
) -> Result<String, InstallableError> {
    validate_component_name(dev_shell)?;

    if !canonical_workspace.is_absolute() || !canonical_flake.is_absolute() {
        return Err(InstallableError::new(
            "canonical workspace and flake paths must be absolute",
        ));
    }

    canonical_flake
        .strip_prefix(canonical_workspace)
        .map_err(|_| {
            InstallableError::new(format!(
                "environment flake `{}` resolves outside workspace `{}`",
                canonical_flake.display(),
                canonical_workspace.display()
            ))
        })?;

    Ok(format!(
        "path:{}#{dev_shell}",
        canonical_flake.to_string_lossy()
    ))
}

pub fn moon_shell_environment(
    canonical_workspace: &Path,
    components: &[String],
) -> Result<BTreeMap<String, String>, InstallableError> {
    if !canonical_workspace.is_absolute() {
        return Err(InstallableError::new(
            "canonical workspace path must be absolute",
        ));
    }

    let serialized_components = components
        .iter()
        .map(|component| {
            validate_component_name(component)?;
            Ok(format!("\"{component}\""))
        })
        .collect::<Result<Vec<_>, InstallableError>>()?
        .join(",");

    Ok(BTreeMap::from([
        (
            MOON_FLAKE_ENV.to_owned(),
            format!("path:{}", canonical_workspace.to_string_lossy()),
        ),
        (
            MOON_COMPONENTS_ENV.to_owned(),
            format!("[{serialized_components}]"),
        ),
    ]))
}
