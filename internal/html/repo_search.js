const search = new URLSearchParams(window.location.search).get("q");
if (search !== "") document.querySelector("#search").value = search;
