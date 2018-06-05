$(document).ready(function(){
    var lastPage = window.localStorage.getItem("lastPage");
    if (!~window.location.pathname.indexOf("/post/")) {
        // If we're not on a post.
        window.localStorage.setItem("lastPage", window.location.pathname);
    }

    $(".back-btn").click(function(){
        window.location.replace(lastPage);
    });
});