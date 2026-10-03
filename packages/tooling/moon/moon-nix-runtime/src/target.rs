/// A resolved flake root and whether it belongs to the task's project.
#[derive(Clone, Debug, Eq, PartialEq)]
pub struct FlakeTarget {
    pub root: String,
    pub is_project_flake: bool,
}

/// Prefer a verified project-flake root, then fall back to the workspace root.
pub fn resolve_flake_target(
    workspace_root: Option<String>,
    project_flake_root: Option<String>,
) -> Option<FlakeTarget> {
    project_flake_root
        .filter(|root| !root.is_empty())
        .map(|root| FlakeTarget {
            root,
            is_project_flake: true,
        })
        .or_else(|| {
            workspace_root
                .filter(|root| !root.is_empty())
                .map(|root| FlakeTarget {
                    root,
                    is_project_flake: false,
                })
        })
}
