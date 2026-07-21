use crate::serialize::quote_posix;
use crate::target::FlakeTarget;

/// A deterministic adapter output derived from explicit host facts and inputs.
#[derive(Clone, Debug, Eq, PartialEq)]
pub enum WrapDecision {
    Unchanged,
    Command { command: String, args: Vec<String> },
    Script(String),
}

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum DevelopTarget {
    Expression(String),
    Installable(String),
}

/// Build a `nix develop` installable from a flake target and optional devShell.
pub fn flake_ref(root: &str, shell: Option<&str>) -> String {
    match shell {
        Some(shell) => format!("{root}#{shell}"),
        None => root.to_owned(),
    }
}

/// Drop a missing named devShell only for a project-owned flake.
pub fn effective_shell(
    target: &FlakeTarget,
    shell: Option<String>,
    named_shell_available: bool,
) -> Option<String> {
    match shell {
        Some(_) if target.is_project_flake && !named_shell_available => None,
        other => other,
    }
}

/// Build the exact argv replacement for a command task.
pub fn plan_command(
    target: Option<&FlakeTarget>,
    shell: Option<&str>,
    command: String,
    child_args: Vec<String>,
) -> WrapDecision {
    let Some(target) = target else {
        return WrapDecision::Unchanged;
    };

    let mut args = vec![
        "develop".to_owned(),
        "--no-update-lock-file".to_owned(),
        flake_ref(&target.root, shell),
        "--command".to_owned(),
        command,
    ];
    args.extend(child_args);

    WrapDecision::Command {
        command: "nix".to_owned(),
        args,
    }
}

/// Build the exact shell wrapper for an opaque script task.
pub fn plan_script(
    target: Option<&FlakeTarget>,
    shell: Option<&str>,
    script: &str,
) -> WrapDecision {
    let Some(target) = target else {
        return WrapDecision::Unchanged;
    };

    WrapDecision::Script(format!(
        "nix develop --no-update-lock-file {} --command bash -c {}",
        quote_posix(&flake_ref(&target.root, shell)),
        quote_posix(script)
    ))
}

/// Build the exact argv replacement for an extension-managed command task.
pub fn plan_develop_command(
    target: &DevelopTarget,
    command: String,
    child_args: Vec<String>,
) -> WrapDecision {
    let mut args = develop_args(target);
    args.extend(["--command".to_owned(), command]);
    args.extend(child_args);

    WrapDecision::Command {
        command: "nix".to_owned(),
        args,
    }
}

/// Build the shell wrapper for an opaque extension-managed script task.
pub fn plan_develop_script(target: &DevelopTarget, script: &str) -> WrapDecision {
    let mut args = vec!["nix".to_owned()];
    args.extend(develop_args(target));
    args.extend([
        "--command".to_owned(),
        "bash".to_owned(),
        "-c".to_owned(),
        script.to_owned(),
    ]);

    WrapDecision::Script(
        args.iter()
            .map(|argument| quote_posix(argument))
            .collect::<Vec<_>>()
            .join(" "),
    )
}

fn develop_args(target: &DevelopTarget) -> Vec<String> {
    let mut args = vec![
        "develop".to_owned(),
        "--impure".to_owned(),
        "--no-update-lock-file".to_owned(),
    ];

    match target {
        DevelopTarget::Expression(expression) => {
            args.extend(["--expr".to_owned(), expression.clone(), "moon".to_owned()])
        }
        DevelopTarget::Installable(installable) => args.push(installable.clone()),
    }

    args
}
