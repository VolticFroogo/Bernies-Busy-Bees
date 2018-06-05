$(document).ready(function(){
    M.AutoInit();
    $('textarea, input').characterCounter();
    Waves.displayEffect();

    $(".thumbnail-btn").click(function(){
        $(".thumbnail").trigger('click');
    });

    $(".images-btn").click(function(){
        $(".images").trigger('click');
    });

    $(".submit-btn").click(function(){
        M.toast({html: "Sending new post request!"});
        var formData = new FormData($(this).parents("form")[0]);

        if (typeof $(".thumbnail")[0].files[0] === "undefined") {
            M.Toast.dismissAll(); // Clear all other toasts.
            M.toast({html: "You need to select a thumbnail."});
            return;
        }

        $.ajax({
            type: "POST",
            url: "/panel/post/new",
            data: formData,
            cache: false,
            contentType: false,
            processData: false,
            success: function(rRaw) {
                var r = JSON.parse(rRaw);
                if (r.success) {
                    window.location.replace(window.localStorage.getItem("lastPage"));
                } else {
                    M.Toast.dismissAll(); // Clear all other toasts.
                    M.toast({html: "There was an error, try again. If this persists, refresh the page."});
                }
            }
        });
    });
});