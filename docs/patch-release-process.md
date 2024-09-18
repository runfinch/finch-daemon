## Finch-Daemon Patch release process

> [!WARNING]
> It is highly recommended to cut new releases from the main branch. Use this process with caution

We create new releases form the main branch and maintain [Conventional Commit messages](https://www.conventionalcommits.org/en/v1.0.0/) in our commits to help build the changelog entries and also automatically determine if the next version release is a major/minor or patch release.

### Scenario: We want to fix a critical bug / CVE  
- **Case 1 : there are no new commits on the mainline other than the fix.** 
    - In this case continue to release a patch release following the normal release process.

- **Case 2 : There are some features on the mainline other than the fix which we do not want to publish yet**
    - Create a new branch from the tip of the last release - `<Release-vX.X.X>`
    - Cherry-pick the fix from the mainline into this branch
    - Update the `release-please.yaml` file to trigger on the current branch name. i.e 
          ``` 
          on:
              push:
                branches:
                  - Release-vX.X.X
            ```

      - Verify that the version number in the `.release-please-manifest.json`  is the expected version. It should be the current released version - release please will increment it when it raises a PR (**OR**) Set the `"release-as":` flag to the desired release version.
      - Update the `“on push”` branch in the CI to the release branch
      - Commit this PR with a `fix:` suffix so that the patch version only changes
    - Merge the changes to the `<Release-vX.X.X>` branch
    - release-please automatically creates a merge PR against this branch
      
      - Close and reopen to run the CI checks 
      - Merge the PR and verify that a new release is created