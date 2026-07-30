use extism_pdk::*;
use moon_nix_runtime::environment::{
    resolve_environment, validate_environment_config, EnvironmentOverride, EnvironmentRequest,
    EnvironmentResolution, OverrideMode,
};
use moon_nix_runtime::guard::{decide_wrap, WrapFacts, WrapGuard, SENTINEL};
use moon_nix_runtime::installable::{
    canonical_installable, moon_shell_environment, parse_workspace_installable,
    MOON_SHELL_EXPRESSION,
};
use moon_nix_runtime::wrap::{
    plan_develop_command, plan_develop_script, DevelopTarget, WrapDecision,
};
use moon_pdk::*;
use moon_pdk_api::*;
use schematic::{Config, DefaultValueResult, SchemaBuilder, ValidateError, ValidateResult};
use std::collections::BTreeMap;
use std::path::{Path, PathBuf};

#[derive(Clone, Config, Debug)]
#[config(rename_all = "camelCase")]
pub struct NixExtensionConfig {
    #[setting(default = default_base_components)]
    pub base_components: Vec<String>,

    #[setting(default = default_environment_by_toolchain)]
    pub environment_by_toolchain: BTreeMap<String, String>,

    #[setting(default = default_environment_overrides, validate = validate_environment_overrides)]
    pub environment_by_project: BTreeMap<String, EnvironmentOverrideConfig>,

    #[setting(default = default_environment_overrides, validate = validate_environment_overrides)]
    pub environment_by_task: BTreeMap<String, EnvironmentOverrideConfig>,

    #[setting(default = true)]
    pub fail_closed: bool,
}

#[derive(Clone, Config, Debug, Eq, PartialEq, serde::Deserialize, serde::Serialize)]
#[config(serde(untagged))]
#[serde(untagged)]
pub enum EnvironmentOverrideConfig {
    #[setting(nested)]
    Components(ComponentOverrideConfig),
    #[setting(nested)]
    Installable(InstallableOverrideConfig),
}

#[derive(Clone, Config, Debug, Eq, PartialEq, serde::Deserialize, serde::Serialize)]
#[config(rename_all = "camelCase")]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct ComponentOverrideConfig {
    #[setting(required)]
    pub components: Option<Vec<String>>,

    #[serde(default)]
    pub mode: OverrideModeConfig,
}

#[derive(Clone, Config, Debug, Eq, PartialEq, serde::Deserialize, serde::Serialize)]
#[config(rename_all = "camelCase")]
#[serde(rename_all = "camelCase", deny_unknown_fields)]
pub struct InstallableOverrideConfig {
    #[setting(required)]
    pub installable: Option<String>,
}

#[derive(Clone, Config, Debug, Eq, PartialEq, serde::Deserialize, serde::Serialize)]
#[config(rename_all = "lowercase")]
#[serde(rename_all = "lowercase")]
pub enum OverrideModeConfig {
    #[setting(default)]
    Append,
    Replace,
}

fn default_base_components(_: &()) -> DefaultValueResult<Vec<String>> {
    Ok(Some(vec!["general".to_owned()]))
}

fn default_environment_by_toolchain(_: &()) -> DefaultValueResult<BTreeMap<String, String>> {
    Ok(Some(BTreeMap::from([
        ("javascript".to_owned(), "node".to_owned()),
        ("node".to_owned(), "node".to_owned()),
        ("npm".to_owned(), "node".to_owned()),
        ("go".to_owned(), "go".to_owned()),
        ("system".to_owned(), "general".to_owned()),
    ])))
}

fn default_environment_overrides(
    _: &(),
) -> DefaultValueResult<BTreeMap<String, EnvironmentOverrideConfig>> {
    Ok(Some(BTreeMap::new()))
}

fn validate_environment_overrides<T>(
    overrides: &BTreeMap<String, EnvironmentOverrideConfig>,
    _: &T,
    _: &(),
    _: bool,
) -> ValidateResult {
    for (scope, environment) in overrides {
        match environment {
            EnvironmentOverrideConfig::Components(config) if config.components.is_none() => {
                return Err(ValidateError::new(format!(
                    "`{scope}` component override requires `components`"
                )));
            }
            EnvironmentOverrideConfig::Installable(config)
                if config
                    .installable
                    .as_deref()
                    .is_none_or(|installable| installable.trim().is_empty()) =>
            {
                return Err(ValidateError::new(format!(
                    "`{scope}` installable override requires a non-empty `installable`"
                )));
            }
            _ => {}
        }
    }

    Ok(())
}

impl From<&EnvironmentOverrideConfig> for EnvironmentOverride {
    fn from(config: &EnvironmentOverrideConfig) -> Self {
        match config {
            EnvironmentOverrideConfig::Components(config) => EnvironmentOverride::Components {
                components: config.components.clone().unwrap_or_default(),
                mode: match config.mode {
                    OverrideModeConfig::Append => OverrideMode::Append,
                    OverrideModeConfig::Replace => OverrideMode::Replace,
                },
            },
            EnvironmentOverrideConfig::Installable(config) => {
                EnvironmentOverride::Installable(config.installable.clone().unwrap_or_default())
            }
        }
    }
}

#[plugin_fn]
pub fn register_extension(
    Json(_): Json<RegisterExtensionInput>,
) -> FnResult<Json<RegisterExtensionOutput>> {
    Ok(Json(RegisterExtensionOutput {
        name: "Nix environment".into(),
        plugin_version: env!("CARGO_PKG_VERSION").into(),
        description: Some(
            "Composes native Moon task toolchains through the workspace Nix flake.".into(),
        ),
    }))
}

#[plugin_fn]
pub fn define_extension_config() -> FnResult<Json<DefineExtensionConfigOutput>> {
    Ok(Json(DefineExtensionConfigOutput {
        schema: SchemaBuilder::build_root::<NixExtensionConfig>(),
    }))
}

#[plugin_fn]
pub fn extend_task_command(
    Json(input): Json<ExtendTaskCommandInput>,
) -> FnResult<Json<ExtendTaskCommandOutput>> {
    let config: NixExtensionConfig = parse_extension_config_schema(input.extension_config.clone())?;
    let (decision, environment) = resolve_command_decision(&input, &config)?;

    Ok(Json(command_output(decision, environment)))
}

#[plugin_fn]
pub fn extend_task_script(
    Json(input): Json<ExtendTaskScriptInput>,
) -> FnResult<Json<ExtendTaskScriptOutput>> {
    let config: NixExtensionConfig = parse_extension_config_schema(input.extension_config.clone())?;
    let (decision, environment) = resolve_script_decision(&input, &config)?;

    Ok(Json(script_output(decision, environment)))
}

fn resolve_command_decision(
    input: &ExtendTaskCommandInput,
    config: &NixExtensionConfig,
) -> AnyResult<(WrapDecision, BTreeMap<String, String>)> {
    let resolution = resolve_task_environment(&input.project, &input.task, config)?;
    let Some((target, environment)) =
        resolve_develop_target(&input.context, &input.task, config, resolution)?
    else {
        return Ok((WrapDecision::Unchanged, BTreeMap::new()));
    };

    Ok((
        plan_develop_command(&target, input.command.clone(), input.args.clone()),
        environment,
    ))
}

fn resolve_script_decision(
    input: &ExtendTaskScriptInput,
    config: &NixExtensionConfig,
) -> AnyResult<(WrapDecision, BTreeMap<String, String>)> {
    let resolution = resolve_task_environment(&input.project, &input.task, config)?;
    let Some((target, environment)) =
        resolve_develop_target(&input.context, &input.task, config, resolution)?
    else {
        return Ok((WrapDecision::Unchanged, BTreeMap::new()));
    };

    Ok((plan_develop_script(&target, &input.script), environment))
}

fn resolve_task_environment(
    project: &ProjectFragment,
    task: &TaskFragment,
    config: &NixExtensionConfig,
) -> AnyResult<EnvironmentResolution> {
    let environment_by_project = runtime_overrides(&config.environment_by_project);
    let environment_by_task = runtime_overrides(&config.environment_by_task);
    validate_environment_config(
        &config.base_components,
        &config.environment_by_toolchain,
        &environment_by_project,
        &environment_by_task,
    )?;
    validate_scoped_keys(config)?;

    let task_target = task.target.as_str();
    let task_toolchains = task
        .toolchains
        .iter()
        .map(|toolchain| toolchain.as_str().to_owned())
        .collect::<Vec<_>>();

    Ok(resolve_environment(EnvironmentRequest {
        base_components: &config.base_components,
        environment_by_toolchain: &config.environment_by_toolchain,
        project_override: environment_by_project.get(project.id.as_str()),
        task_override: environment_by_task.get(task_target),
        task_toolchains: &task_toolchains,
    })?)
}

fn runtime_overrides(
    overrides: &BTreeMap<String, EnvironmentOverrideConfig>,
) -> BTreeMap<String, EnvironmentOverride> {
    overrides
        .iter()
        .map(|(scope, environment)| (scope.clone(), environment.into()))
        .collect()
}

fn validate_scoped_keys(config: &NixExtensionConfig) -> AnyResult<()> {
    for toolchain in config.environment_by_toolchain.keys() {
        Id::new(toolchain).map_err(|error| {
            anyhow!("invalid environmentByToolchain key `{toolchain}`: {error}")
        })?;
    }

    for project_id in config.environment_by_project.keys() {
        Id::new(project_id)
            .map_err(|error| anyhow!("invalid environmentByProject key `{project_id}`: {error}"))?;

        let project = load_project(project_id).map_err(|error| {
            anyhow!("environmentByProject references unknown project `{project_id}`: {error}")
        })?;
        if project.id.as_str() != project_id {
            return Err(anyhow!(
                "environmentByProject references unknown project `{project_id}`"
            ));
        }
    }

    for target in config.environment_by_task.keys() {
        validate_full_target(target)?;

        let task = load_task(target).map_err(|error| {
            anyhow!("environmentByTask references unknown task `{target}`: {error}")
        })?;
        if task.target.as_str() != target {
            return Err(anyhow!(
                "environmentByTask references unknown task `{target}`"
            ));
        }
    }

    Ok(())
}

fn validate_full_target(target: &str) -> AnyResult<()> {
    let Some((project_id, task_id)) = target.split_once(':') else {
        return Err(anyhow!(
            "invalid environmentByTask key `{target}`; expected a full `project:task` target"
        ));
    };

    if project_id.is_empty() || task_id.is_empty() || task_id.contains(':') {
        return Err(anyhow!(
            "invalid environmentByTask key `{target}`; expected a full `project:task` target"
        ));
    }

    Id::new(project_id)
        .map_err(|error| anyhow!("invalid environmentByTask project in `{target}`: {error}"))?;
    Id::new(task_id)
        .map_err(|error| anyhow!("invalid environmentByTask task in `{target}`: {error}"))?;

    Ok(())
}

fn resolve_develop_target(
    context: &MoonContext,
    task: &TaskFragment,
    config: &NixExtensionConfig,
    resolution: EnvironmentResolution,
) -> AnyResult<Option<(DevelopTarget, BTreeMap<String, String>)>> {
    if resolution == EnvironmentResolution::Unchanged {
        return Ok(None);
    }

    let facts = WrapFacts {
        in_nix_shell: !get_host_env_var("IN_NIX_SHELL")?
            .unwrap_or_default()
            .is_empty(),
        already_wrapped: get_host_env_var(SENTINEL)?.unwrap_or_default() == "1",
        nix_available: command_exists(&get_host_environment()?, "nix"),
    };

    match decide_wrap(facts) {
        WrapGuard::Unchanged => return Ok(None),
        WrapGuard::MissingNix if config.fail_closed => {
            return Err(anyhow!(
                "nix is required for `{}` because its toolchains or scoped override activate moon_nix_extension, but `nix` was not found on PATH",
                task.target.as_str()
            ));
        }
        WrapGuard::MissingNix => return Ok(None),
        WrapGuard::Ready => {}
    }

    let canonical_workspace = canonical_workspace_root(context)?;

    Ok(Some(match resolution {
        EnvironmentResolution::Unchanged => unreachable!(),
        EnvironmentResolution::Components(components) => (
            DevelopTarget::Expression(MOON_SHELL_EXPRESSION.to_owned()),
            moon_shell_environment(&canonical_workspace, &components)?,
        ),
        EnvironmentResolution::Installable(installable) => (
            DevelopTarget::Installable(resolve_workspace_installable(
                &canonical_workspace,
                &installable,
            )?),
            BTreeMap::new(),
        ),
    }))
}

fn canonical_workspace_root(context: &MoonContext) -> AnyResult<PathBuf> {
    let workspace_root = context
        .workspace_root
        .real_path()
        .ok_or_else(|| anyhow!("Moon workspace root has no real filesystem path"))?;

    canonicalize_host_path(&workspace_root, "workspace")
}

fn resolve_workspace_installable(
    canonical_workspace: &Path,
    installable: &str,
) -> AnyResult<String> {
    let parsed = parse_workspace_installable(installable)?;
    let flake_candidate = canonical_workspace.join(parsed.relative_flake);
    let canonical_flake = canonicalize_host_path(&flake_candidate, "environment flake")?;
    let canonical_reference =
        canonical_installable(canonical_workspace, &canonical_flake, &parsed.dev_shell)?;

    for required_file in ["flake.nix", "flake.lock"] {
        let required_path = canonical_flake.join(required_file);
        let required_path_text = required_path.to_string_lossy();

        if !exec_captured("test", ["-f", required_path_text.as_ref()])
            .is_ok_and(|result| result.exit_code == 0)
        {
            return Err(anyhow!(
                "environment flake `{}` has no {required_file}",
                canonical_flake.display()
            ));
        }
    }

    Ok(canonical_reference)
}

fn canonicalize_host_path(path: &Path, label: &str) -> AnyResult<PathBuf> {
    let path_text = path.to_string_lossy();
    let result = exec_captured(
        "sh",
        [
            "-c",
            "CDPATH= cd -P \"$1\" && pwd -P",
            "moon-nix-extension",
            path_text.as_ref(),
        ],
    )
    .map_err(|error| {
        anyhow!(
            "failed to canonicalize {label} path `{}`: {error}",
            path.display()
        )
    })?;

    if result.exit_code != 0 {
        return Err(anyhow!(
            "failed to canonicalize {label} path `{}`: {}",
            path.display(),
            result.stderr.trim()
        ));
    }

    Ok(PathBuf::from(result.stdout.trim()))
}

fn command_output(
    decision: WrapDecision,
    environment: BTreeMap<String, String>,
) -> ExtendTaskCommandOutput {
    let mut output = ExtendTaskCommandOutput::default();
    let WrapDecision::Command { command, args } = decision else {
        return output;
    };

    output.command = Some(command);
    output.args = Some(Extend::Replace(args));
    output.env.extend(environment);
    output.env.insert(SENTINEL.into(), "1".into());
    output
}

fn script_output(
    decision: WrapDecision,
    environment: BTreeMap<String, String>,
) -> ExtendTaskScriptOutput {
    let mut output = ExtendTaskScriptOutput::default();
    let WrapDecision::Script(script) = decision else {
        return output;
    };

    output.script = Some(script);
    output.env.extend(environment);
    output.env.insert(SENTINEL.into(), "1".into());
    output
}
