---
name: gumlet-video-workspace
description: Manage Gumlet video workspaces using the Gumlet CLI. Use this skill to list all workspaces, get details of a specific workspace, create a new workspace, rename an existing workspace, or delete a workspace. Requires the Gumlet CLI to be installed and authenticated via `gumlet login`.
---

## List all workspaces

```sh
gumlet video workspace list
gumlet video workspace list --output table
```

## Get details of a workspace

```sh
gumlet video workspace get --workspace-id <id>
```

## Create a new workspace

```sh
gumlet video workspace create --name "My Workspace"
```

## Rename a workspace

```sh
gumlet video workspace update --workspace-id <id> --name "New Name"
```

## Delete a workspace

```sh
gumlet video workspace delete --workspace-id <id>
```

## Notes

- `-o table` shows a human-readable table. Default JSON includes the unmodified API fields and any pagination metadata in the response.
- The `--workspace-id` flag is required for `get`, `update`, and `delete`.
