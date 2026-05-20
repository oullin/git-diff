# Cask template for the Homebrew tap at oullin/homebrew-tap.
#
# This file is the source of truth; the release pipeline copies it to the tap
# repo and substitutes ${version} and ${sha256} from the freshly published
# GitHub release artifact.

cask "git-diff" do
  version "0.0.0"
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"

  url "https://github.com/oullin/git-diff/releases/download/v#{version}/git-diff-review-#{version}-arm64.dmg"
  name "Git Diff Review"
  desc "Local git diff reviewer"
  homepage "https://github.com/oullin/git-diff"

  app "Git Diff Review.app"

  zap trash: [
    "~/Library/Application Support/git-diff",
    "~/Library/Preferences/io.gocanto.git-diff.plist",
    "~/Library/Saved Application State/io.gocanto.git-diff.savedState",
  ]
end
