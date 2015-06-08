<!DOCTYPE html>
<html>
<head>
    <title>用户注册</title>
    <link href="/public/stylesheets/bootstrap.min.css" rel="stylesheet" media="screen">
</head>
<body screen_capture_injected="true">
<div class="container-fluid">
    <form class="form-horizontal" method="post" action="/reg">
        <fieldset>
            <legend>用户注册</legend>
            {{if .Msg}}
            <div class="alert alert-error">{{.Msg}}</div>
            {{end}}
            <div class="control-group">
                <label class="control-label" for="user">用户名</label>
                <div class="controls">
                    <input type="text" class="input-xlarge" id="user" name="user" required="true" autofocus="true">
                </div>
            </div>
            <div class="control-group">
                <label class="control-label" for="user">密码</label>
                <div class="controls">
                    <input type="password" class="input-xlarge" id="pwd" name="pwd" required="true">
                </div>
            </div>
            <div class="control-group">
                <label class="control-label" for="user">确认密码</label>
                <div class="controls">
                    <input type="password" class="input-xlarge" id="pwd2" name="pwd2" required="true">
                </div>
            </div>
            <div class="form-actions">
                <button type="submit" class="btn btn-primary">注册</button>
                <a href="/login">登陆</a>
            </div>
        </fieldset>
    </form>
</div>
<script src="/public/javascripts/jquery-1.7.2.min.js"></script>
<script src="/public/javascripts/bootstrap.min.js"></script>

<script type="text/javascript">
    $(".form-horizontal").submit(function(event) {
        if( $("#user").val() == '' ) {
            alert("请输入用户名");
            $("#user").focus();
            return false;
        }
        if( $("#pwd").val() == '' ) {
            alert("请输入密码");
            $("#pwd").focus();
            return false;
        }
        if( $("#pwd").val().length < 6 || $("#pwd").val().length > 20 ) {
            alert("密码只能在6-20位之间");
            $("#pwd").focus();
            return false;
        }
        if( $("#pwd").val() != $("#pwd2").val() ) {
            alert("两次输入的密码不一致");
            $("#pwd").focus();
            return false;
        }
    });
</script>
</body>
</html>