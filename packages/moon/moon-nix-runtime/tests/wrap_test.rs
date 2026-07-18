use moon_nix_runtime::target::FlakeTarget;
use moon_nix_runtime::wrap::{
    effective_shell, plan_command, plan_develop_command, plan_develop_script, plan_script,
    DevelopTarget, WrapDecision,
};

fn workspace_target() -> FlakeTarget {
    FlakeTarget {
        root: "/workspace".to_owned(),
        is_project_flake: false,
    }
}

#[test]
fn command_plan_preserves_the_exact_child_argv() {
    let target = workspace_target();

    let decision = plan_command(
        Some(&target),
        Some("go"),
        "golangci-lint".to_owned(),
        vec!["run".to_owned(), "--fix".to_owned()],
    );

    assert_eq!(
        decision,
        WrapDecision::Command {
            command: "nix".to_owned(),
            args: vec![
                "develop".to_owned(),
                "/workspace#go".to_owned(),
                "--command".to_owned(),
                "golangci-lint".to_owned(),
                "run".to_owned(),
                "--fix".to_owned(),
            ],
        }
    );
}

#[test]
fn script_plan_quotes_the_flake_and_opaque_script_exactly() {
    let target = workspace_target();

    let decision = plan_script(Some(&target), Some("shell"), "echo 'hi'");

    assert_eq!(
        decision,
        WrapDecision::Script(
            r#"nix develop '/workspace#shell' --command bash -c 'echo '\''hi'\'''"#.to_owned()
        )
    );
}

#[test]
fn expression_command_plan_preserves_the_exact_child_argv() {
    let expression = "let flake = builtins.getFlake \"path:/workspace\"; in { moon = flake.lib.mkMoonShell builtins.currentSystem [ \"general\" ]; }";

    let decision = plan_develop_command(
        &DevelopTarget::Expression(expression.to_owned()),
        "node".to_owned(),
        vec!["script with spaces.js".to_owned(), "${literal}".to_owned()],
    );

    assert_eq!(
        decision,
        WrapDecision::Command {
            command: "nix".to_owned(),
            args: vec![
                "develop".to_owned(),
                "--impure".to_owned(),
                "--expr".to_owned(),
                expression.to_owned(),
                "moon".to_owned(),
                "--command".to_owned(),
                "node".to_owned(),
                "script with spaces.js".to_owned(),
                "${literal}".to_owned(),
            ],
        }
    );
}

#[test]
fn installable_script_plan_quotes_every_argument_and_keeps_the_script_opaque() {
    let decision = plan_develop_script(
        &DevelopTarget::Installable("path:/workspace with spaces/project#go".to_owned()),
        "printf '%s\\n' \"$HOME\"",
    );

    assert_eq!(
        decision,
        WrapDecision::Script(
            "'nix' 'develop' '--impure' 'path:/workspace with spaces/project#go' '--command' 'bash' '-c' 'printf '\\''%s\\n'\\'' \"$HOME\"'"
                .to_owned()
        )
    );
}

#[test]
fn absent_target_keeps_a_command_unchanged() {
    let decision = plan_command(None, None, "echo".to_owned(), vec!["hi".to_owned()]);

    assert_eq!(decision, WrapDecision::Unchanged);
}

#[test]
fn missing_project_shell_falls_back_to_default() {
    let target = FlakeTarget {
        root: "/workspace/project".to_owned(),
        is_project_flake: true,
    };

    let shell = effective_shell(&target, Some("go".to_owned()), false);

    assert_eq!(shell, None);
}

#[test]
fn workspace_shell_does_not_require_a_project_probe() {
    let target = workspace_target();

    let shell = effective_shell(&target, Some("go".to_owned()), false);

    assert_eq!(shell.as_deref(), Some("go"));
}
