const lineRe = /#L(\d+)(?:-L(\d+))?/g
const $lineLines = document.querySelectorAll(".chroma .lntable .lnt");
const $codeLines = document.querySelectorAll(".chroma .lntable .line");
const $copyButton = document.getElementById('copy');
const $permalink = document.getElementById('permalink');
const $copyIcon = "📋";
const $copiedIcon = "✅";
let $code = ""
for (let codeLine of $codeLines) $code += codeLine.innerText;
let start = 0;
let end = 0;
const results = [...location.hash.matchAll(lineRe)];
if (0 in results) {
    start = results[0][1] !== undefined ? parseInt(results[0][1]) : 0;
    end = results[0][2] !== undefined ? parseInt(results[0][2]) : 0;
}
if (start !== 0) {
    deactivateLines();
    activateLines(start, end);
    let anchor = `#${start}`;
    if (end !== 0) anchor += `-${end}`;
    if (anchor !== "") $permalink.href = $permalink.dataset.permalink + anchor;
    $lineLines[start - 1].scrollIntoView(true);
}
for (let line of $lineLines) {
    line.addEventListener("click", (event) => {
        event.preventDefault();
        deactivateLines();
        const n = parseInt(line.id.substring(1));
        let anchor = "";
        if (event.shiftKey) {
            end = n;
            anchor = `#L${start}-L${end}`;
        } else if (start === n) {
            start = 0;
            end = 0;
        } else {
            start = n;
            end = 0;
            anchor = `#L${start}`;
        }
        history.replaceState(null, null, window.location.pathname + anchor);
        $permalink.href = $permalink.dataset.permalink + anchor;
        if (start !== 0) activateLines(start, end);
    });
}
if (navigator.clipboard && navigator.clipboard.writeText) {
    $copyButton.innerText = $copyIcon;
    $copyButton.classList.remove("hidden");
}
$copyButton.addEventListener("click", () => {
    navigator.clipboard.writeText($code);
    $copyButton.innerText = $copiedIcon;
    setTimeout(() => {
        $copyButton.innerText = $copyIcon;
    }, 1000);
});

function activateLines(start, end) {
    if (end < start) end = start;
    for (let idx = start - 1; idx < end; idx++) {
        $codeLines[idx].classList.add("active");
    }
}

function deactivateLines() {
    for (let code of $codeLines) {
        code.classList.remove("active");
    }
}
