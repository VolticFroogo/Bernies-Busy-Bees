$(document).ready(function(){
    M.AutoInit();
    $('textarea, input').characterCounter();
    Waves.displayEffect();

    var titleContent = $('#title').html();
    $('#title').blur(function() {
        if (titleContent!==$(this).html()){
            titleContent = $(this).html();
            updatePost();
        }
    });

    var descriptionContent = $('#title').html();
    $('#description').blur(function() {
        if (descriptionContent!==$(this).html()){
            descriptionContent = $(this).html();
            updatePost();
        }
    });

    function updatePost() {
        $.ajax({
            url: "/panel/post/update",
            type: "post",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                ID: PostID,
                Title: titleContent,
                Description: descriptionContent,
                CsrfSecret: CsrfSecret
            }),
            dataType: "json",
            success: function(r) {
                if(r.success) {
                    document.title = "BBB | " + titleContent;
                } else {
                    M.Toast.dismissAll(); // Clear all other toasts.
                    M.toast({html: "Error updating post, refresh the page."});
                }
            }
        });
    }

    $(document).keypress(function(event){
        if (event.keyCode === 10 || event.keyCode === 13) 
            event.preventDefault();
    });

    $("#delete-btn").click(function(){
        M.toast({html: "Deleting post."});

        $.ajax({
            url: "/panel/post/delete",
            type: "post",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                ID: PostID,
                CsrfSecret: CsrfSecret
            }),
            dataType: "json",
            success: function(r) {
                if(r.success) {
                    window.location.replace(window.localStorage.getItem("lastPage"));
                } else {
                    M.Toast.dismissAll(); // Clear all other toasts.
                    M.toast({html: "Error deleting post, refresh the page."});
                }
            }
        });
    });

    $("#comment-btn").click(function(){
        var comment = $('#comment-textarea').val();

        if (comment === "") {
            M.toast({html: "You need to enter a comment first."});
            return;
        }

        M.toast({html: "Submitting comment."});

        $.ajax({
            url: "/panel/post/comment",
            type: "post",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                ID: PostID,
                Comment: comment,
                CsrfSecret: CsrfSecret
            }),
            dataType: "json",
            success: function(r) {
                M.Toast.dismissAll(); // Clear all other toasts.
                if(r.success) {
                    var commentID = r.id;
                    $("#comment-section").append('<div class="col s12 comment"> <div class="card-panel grey lighten-5 z-depth-1 hoverable"> <div style="font-size: 140%;">' + Fname + ' ' + Lname + '</div><span>Just now</span> <p>' + comment + '</p><a class="delete-comment-btn btn-floating waves-effect waves-light red right" style="top: -30px; right: -5px;" data-id="' + commentID + '"><i class="material-icons">delete</i></a> </div></div>');
                    $('#comment-textarea').val(""); // Make the comment box empty.
                    M.toast({html: "Successfully added comment!"});
                } else {
                    M.toast({html: "Error submitting comment, refresh the page."});
                }
            }
        });
    });

    $("#comment-section").on("click", ".delete-comment-btn", function(){
        M.toast({html: "Deleting comment."});
        var id = $(this).attr("data-id");
        var comment = $(this).closest(".comment");

        $.ajax({
            url: "/admin/user/delete",
            type: "POST",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                CsrfSecret: CsrfSecret,
                ID: ID
            }),
            dataType: "json",
            success: function(r) {
                M.Toast.dismissAll(); // Clear all other toasts.
                if (r.success) {
                    $(user).remove();
                    M.toast({html: "Successfully deleted comment!"});
                } else {
                    M.toast({html: "Error deleting comment, refresh the page."});
                }
            }
        });
    });
});


function TimeAgo(current, previous) {
    var secPerMinute = 60;
    var secPerHour = secPerMinute * 60;
    var secPerDay = secPerHour * 24;
    var secPerMonth = secPerDay * 30;
    var secPerYear = secPerDay * 365;
    
    var elapsed = current - previous;
    
    if (elapsed < secPerMinute) {
        return "Just now";
    } else if (elapsed < secPerHour) {
        return timeDifference(elapsed, secPerMinute, "minute");
    } else if (elapsed < secPerDay ) {
        return timeDifference(elapsed, secPerHour, "hour");
    } else if (elapsed < secPerMonth) {
        return timeDifference(elapsed, secPerDay, "day");
    } else if (elapsed < secPerYear) {
        return timeDifference(elapsed, secPerMonth, "month");
    } else {
        return timeDifference(elapsed, secPerYear, "year");
    }
}

function timeDifference(elapsed, period, periodName) {
    var rounded = Math.round(elapsed/period);

    if (rounded === 1) {
        return rounded + " " + periodName + " ago";
    }

    return rounded + " " + periodName + "s ago";
}