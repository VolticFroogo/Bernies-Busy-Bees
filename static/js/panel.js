$(document).ready(function(){
    M.AutoInit();
    Waves.displayEffect();
    $('select').formSelect();

    // User Update
    $("#users").on("click", ".user-update", function(){
        var user = $(this).closest(".user-li");
        var local = parseInt(user.attr("data-local"));

        if (local === 1) {
            UserNew(user);
            return;
        }

        var id = parseInt(user.attr("data-id"));

        M.toast({html: "Updating user."});
    
        $.ajax({
            url: "/panel/user/update",
            type: "POST",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                CsrfSecret: CsrfSecret,
                ID: id,
                Email: user.find(".user-email").val(),
                Password: user.find(".user-password").val(),
                Fname: user.find(".user-fname").val(),
                Lname: user.find(".user-lname").val(),
                Privileges: parseInt(user.find(".user-privileges").val())
            }),
            dataType: "json",
            success: function(r) {
                M.Toast.dismissAll(); // Clear all other toasts.
                if (r.success) {
                    user.find(".user-header").text(user.find(".user-fname").val() + " " + user.find(".user-lname").val());
                    M.toast({html: "Successfully updated user."});
                } else {
                    M.toast({html: "Error updating user, refresh the page."});
                }
            }
        });
    });

    // User New
    function UserNew(user) {
        M.toast({html: "Adding new user."});
    
        $.ajax({
            url: "/panel/user/new",
            type: "POST",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                CsrfSecret: CsrfSecret,
                Email: user.find(".user-email").val(),
                Password: user.find(".user-password").val(),
                Fname: user.find(".user-fname").val(),
                Lname: user.find(".user-lname").val(),
                Privileges: parseInt(user.find(".user-privileges").val())
            }),
            dataType: "json",
            success: function(r) {
                M.Toast.dismissAll(); // Clear all other toasts.
                if (r.success) {
                    user.attr("data-local", "0");
                    user.find(".user-header").text(user.find(".user-fname").val() + " " + user.find(".user-lname").val());
                    user.attr("data-id", r.id);
                    M.toast({html: "Successfully added new user."});
                } else {
                    M.toast({html: "Error adding new user, refresh the page."});
                }
            }
        });
    }

    // User Delete
   $("#users").on("click", ".user-delete", function(){
        M.toast({html: "Deleting user."});

        var user = $(this).closest(".user-li");
        var id = parseInt(user.attr("data-id"));
        var local = parseInt(user.attr("data-local"));

        if (local === 1) {
            user.remove();
            M.Toast.dismissAll(); // Clear all other toasts.
            M.toast({html: "Successfully deleted user."});
            return;
        }

        $.ajax({
            url: "/panel/user/delete",
            type: "POST",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                CsrfSecret: CsrfSecret,
                ID: id
            }),
            dataType: "json",
            success: function(r) {
                M.Toast.dismissAll(); // Clear all other toasts.
                if (r.success) {
                    user.remove();
                    M.toast({html: "Successfully deleted user."});
                } else {
                    M.toast({html: "Error deleting user, refresh the page."});
                }
            }
        });
    });

    // User Add
    $("#user-add").click(function() {
        $("#users ul").append('<li class="user-li" data-local="1"> <div class="collapsible-header user-header">New User</div> <div class="collapsible-body"><span> <div class="row"> <div class="input-field col s12"> <input class="user-email" type="text" data-length="256" maxlength="256"> <label>Email</label> </div> <div class="input-field col s12"> <input class="user-password tooltipped" data-position="top" data-delay="50" data-tooltip="You can leave the password blank to not change it." type="password" data-length="64" maxlength="64"> <label>Password</label> </div> <div class="input-field col s12"> <input class="user-fname" type="text" data-length="16" maxlength="16"> <label>First Name</label> </div> <div class="input-field col s12"> <input class="user-lname" type="text" data-length="16" maxlength="16"> <label>Last Name</label> </div> <div class="input-field col s12"> <select class="user-privileges" autocomplete="off"> <option value="1" selected>Parent</option> <option value="2">Moderator</option> <option value="3">Admin</option> </select> <label>Privileges</label> </div> <div class="input-field col"> <a class="btn waves-effect waves-light purple darken-3 user-update">Submit<i class="material-icons right">send</i></a> <a class="btn waves-effect waves-light red user-delete">Delete<i class="material-icons right">delete</i></a> </div> </div> </span></div></li>');
        $("select").formSelect();
        $('.collapsible').collapsible('open', $('#users ul .user-li').length - 1);
    });

    // Settings Update
    var email = $(".email").val();
    $("#update-settings").click(function() {
        M.toast({html: "Updating your settings."});

        $.ajax({
            url: "/panel/settings/update",
            type: "POST",
            contentType: "application/json; charset=utf-8",
            data: JSON.stringify({
                CsrfSecret: CsrfSecret,
                Email: $(".email").val(),
                Password: $(".password").val(),
                Fname: $(".fname").val(),
                Lname: $(".lname").val()
            }),
            dataType: "json",
            success: function(r) {
                M.Toast.dismissAll(); // Clear all other toasts.
                if (r.success) {
                    M.toast({html: "Successfully updated your settings."});
                    if (email !== $(".email").val()) {
                        M.toast({html: "Please check your email for a verification message."});
                    }
                } else {
                    M.toast({html: "Error updating your settings, refresh the page."});
                }
            }
        });
    });
});