`sqlc-symlink.tar.gz` is a deterministic USTAR archive with one symbolic-link
member, `sqlc` pointing to `extra`. Timestamps and numeric owners are zero.
It tests archive-member rejection without filesystem symlink privileges or
Git Bash link emulation. This first-party fixture uses the repository MIT license.
