## Join Sr Golang/Postgres Developer Take-Home

Hey there, here's the code for the code review take-home. You've got a couple of options:

### 1. Text-Only

1. Start a new, bare private repo.
1. In your new repo, open a pull request that adds the code in this repo.
   - This is you acting as the hypothetical Author of the code.
1. Leave comments against the pull request.
   - This is you acting as yourself - the code reviewer.
1. Invite me to the repo when you're done.

### 2. Video

1. Take this repo's code and put it into whatever format you prefer. Locally-cloned, in a newly-forked repo, or right here.
1. Record yourself (Loom is a good option) walking through the code suggesting alternatives.
1. Send a link to the video to me when you're done.


## Summary of changes
- Add github actions to lint and build the code
- Add goreleaser to build the application
- Move main to `cmd/`
- Move modules to `pkg/`
- Add `vendor` packages
- Synchronize deps
- Add `.gitignore`
- Add configuration
- Add structured logging
- Add pprof
