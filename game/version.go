package game

// Version is the released version of this repository. tagpr rewrites it
// in the release pull request so that it always matches the git tag (see
// .tagpr); it is not edited by hand.
//
// It exists because the board has neither a clock nor a version display,
// so a constant compiled into the firmware is the only handle for
// telling later which source a flashed board is running. On main between
// releases it names the last release, not the working tree.
//
// It lives in this package because both frontends already import it,
// not because it says anything about the game.
const Version = "0.0.0"
