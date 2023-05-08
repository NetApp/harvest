---
name: Harvest Release
about: Use this template for a new Harvest release
title: 'Release [version] (like 23.02.0)'
labels: release
---

### The week before the release

- [ ] Ensure CI is green
- [ ] Create release branch from main to freeze commits, like so:
```bash
RELEASE=23.02.0
git fetch origin
git checkout origin/HEAD
git switch --create release/$RELEASE
git push origin release/$RELEASE
```
- [ ] Create a release branch for the harvest-metrics repo like so:
```bash
RELEASE=23.02.0
git clone https://github.com/NetApp/harvest-metrics.git
git checkout origin/HEAD
git switch --create release/$RELEASE
git push origin release/$RELEASE
```
- [ ] Ensure all issues for the release are tagged with `status/testme` and unassigned. Use `gh` or the GitHub UI to do this.
- [ ] Use the release [issue burn down list](https://github.com/NetApp/harvest/issues?q=is%3Aissue+label%3Astatus%2Ftestme%2Cstatus%2Fopen+sort%3Acreated-asc) to verify issues are fixed. Move `status/testme` issues to `status/open` or `status/done`
- [ ] Ensure that the release is validated against NABox.
- [ ] Ensure that the release is validated against FSX.
- [ ] Use [Jenkins](https://github.com/NetApp/harvest-private/wiki/Release-Checklist#jenkins) to create release artifacts for test machines
- [ ] Create changelog
  - [ ] [Draft a new release](https://github.com/NetApp/harvest/releases). Use `v$RELEASE` for the tag and pick the release/$RELEASE branch. Click the `Generate release notes` button and double check, at the bottom of the release notes, that the commits are across the correct range. For example: `https://github.com/NetApp/harvest/compare/v22.11.1...v23.02.0`
  - [ ] Copy/paste the generated release notes and save them in a file `pbpaste > ghrn_$RELEASE.md`
  - [ ] Hand-write list of release highlights `vi highlights_$RELEASE.md` ([example content](https://github.com/NetApp/harvest/blob/main/CHANGELOG.md#23020--2023-02-21))
    - [ ] Ensure all notable features are highlighted
    - [ ] Ensure any breaking changes are highlighted
    - [ ] Ensure any deprecations are highlighted
  - [ ] Generate changelog by running 
```bash
go run pkg/changelog/main.go --title $RELEASE --highlights releaseHighlights_$RELEASE.md -r ghrn_$RELEASE.md | pbcopy
```
  - [ ] Open a PR against the release branch with the generated release notes for review
  - [ ] PR approval
- [ ] Update metrics repo if needed

#### Update Metrics Documentation
```bash
bin/harvest generate metrics
```
- [ ] Make sure docs look good and open a PR for review with `docs/ontap-metrics.md` changes

### The day of the release

- [ ] Create a new build from [Jenkins](http://harvest-jenkins.rtp.openenglab.netapp.com:8080/job/harvest2_0/job/BuildHarvestArtifacts/) ([details](https://github.com/NetApp/harvest-private/wiki/Release-Checklist#jenkins))
  - [ ] Click `Build with Parameters` and fill in the appropriate fields. Here's an example, where `RELEASE=23.02.0`

| Field                       | Value           |
|-----------------------------|-----------------|
| VERSION                     | 23.02.0         |
| RELEASE                     | 1               |
| BRANCH                      | release/23.02.0 |
| ASUP_MAKE_TARGET            | production      |
| DOCKER_PUBLISH              | true            |
| RUN_TEST                    | true            |
| OVERWRITE_DOCKER_LATEST_TAG | true            |

- [ ] [Draft a new release](https://github.com/NetApp/harvest/releases). Use `v$RELEASE` for the tag and pick the release/$RELEASE branch.
- [ ] Type `$RELEASE` in the `Release title` text input 
- [ ] Open the `CHANGELOG.md` file, copy the single $RELEASE section at the top, and paste into the release notes text area. 
- [ ] Ensure the `Set as the latest release` checkbox is selected
- [ ] Ensure the `Create a discussion for this release ` checkbox is selected
- [ ] Upload the artifacts from the Jenkins build above and attach to the release
- [ ] Click `Publish release` button
- [ ] Announce on Discord 
- [ ] Publish latest docs to netapp.github.io
  - [ ] Make sure you're on the release branch `git switch release/$RELEASE`
  - [ ] The documentation version selector uses the short form of the release, so `23.02` NOT `23.02.0`
```bash
mike deploy --push --update-aliases $SHORT latest
```
- [ ] Merge Release Branch into Main
