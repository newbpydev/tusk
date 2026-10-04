# PR #7 HTML target-policy correction

The [receipt](acceptance.json) binds the owner-approved Goldmark follow-up to
Red/Green controls and official Go 1.27.1 canonical validation.

The immutable `87a9396` baseline accepts slash-separated raw image attributes
and remote Markdown image text inside HTML blocks. The correction retains
explicit raw URL rules and conservatively inspects every Goldmark HTML block
kind, closing lines and inline raw HTML. This preserves source-policy refusal;
it does not claim Markdown inside raw HTML renders as an image.

Approved URLs, ordinary HTML content, CDATA markers and supported code excerpts
remain passing controls. The class audit also reproduced refusal of approved
unquoted URLs after slash separators. Complete unquoted extraction now applies
the same boundaries to `src`, `href` and `srcset`.

The first canonical run passed before the extra unquoted controls and is retained
as superseded evidence. The final marked canonical run validates the complete
correction, passing 739 named script assertions and 98.1% helper coverage. Input hashes distinguish pre-command source/test/authority inputs
from final outcome-only authority updates. Hosted settlement and final release
acceptance remain pending at this local checkpoint.
