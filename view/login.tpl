<!DOCTYPE html>
<html>
<head>
    <title>用户登陆</title>
    <link href="/public/stylesheets/bootstrap.min.css" rel="stylesheet" media="screen">
</head>
<body screen_capture_injected="true">
<div class="container-fluid">
    <form class="form-horizontal" method="post" action="/login">
        <fieldset>
            <legend>用户登陆</legend>
            {{if .Msg}}
            <div class="alert alert-error">{{.Msg}}</div>
            {{end}}
            <div class="control-group">
                <label class="control-label" for="user">用户名</label>
                <div class="controls">
                    <input type="text" class="input-xlarge" id="user" name="user" required="true" />
                </div>
            </div>
            <div class="control-group">
                <label class="control-label" for="pwd">密码</label>
                <div class="controls">
                    <input type="password" class="input-xlarge" id="pwd" name="pwd" required="true" />
                </div>
            </div>
            <div class="form-actions">
                <button type="submit" class="btn btn-primary">登陆</button>
                <a href="/reg">注册</a>
            </div>
        </fieldset>
    </form>
</div>
<script src="/public/javascripts/jquery-1.7.2.min.js"></script>
<script src="/public/javascripts/bootstrap.min.js"></script>
</body>
</html>