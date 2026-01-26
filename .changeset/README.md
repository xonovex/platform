# Changesets

This directory is used by [Changesets](https://github.com/changesets/changesets) to manage versioning and changelogs.

## Adding a Changeset

When you make a change that should be released, run:

```bash
npm run changeset:add
```

This will prompt you to:
1. Select which packages have changed
2. Choose the semver bump type (major/minor/patch)
3. Write a summary of the changes

## Release Process

1. Create PR with your changes and a changeset file
2. CI validates the changeset exists
3. After merge to main, Changesets bot creates a "Version Packages" PR
4. Merging the Version Packages PR triggers automatic publishing
