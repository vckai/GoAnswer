<html>
<head>
<meta charset="utf-8">
<title>!---!</title>
<script type="text/javascript">
  var host = "{{.Host}}";
</script>
<script type="text/javascript" src="/public/javascripts/jquery.min.js"></script>
<script type="text/javascript" src="/public/javascripts/jquery.cookie.js"></script>
<script src="/public/javascripts/bootstrap.min.js"></script>
<script type="text/javascript" src="/public/javascripts/thinkbox/jquery.ThinkBox.min.js"></script>
<script type="text/javascript" src="/public/javascripts/server.js"></script>

<link rel="stylesheet" type="text/css" href="/public/stylesheets/style.css"  />
<link rel="stylesheet" type="text/css" href="/public/stylesheets/bootstrap.css"  />
<link rel="stylesheet" type="text/css" href="/public/javascripts/thinkbox/thinkbox.css"  />

</head>
<body>
<div id="load"><img src="/public/images/loading.gif" />正在连接服务器...</div>

<a id="joinRoom" href="javascript:;">进入房间</a>   <a id="joinRebotRoom" href="javascript:;">进入机器人房间</a>
<a id="outRoom" href="javascript:;">退出房间</a>

<a href="/logout">退出</a>

<div id="main">
    <div class="a">
        <div id="users">  <!-- 用户列表 -->
        <div class="row-fluid">
            <ul class="thumbnails">
            </ul>
          </div>
        </div>
        <div id="timer">50</div>
    </div>
    <legend></legend>
    <div id="examMain">
        <div id="examTitle" class="text-center"></div>
        <div id="examOption"></div>
    </div>   <!-- 答题处理 -->
</div>

</body>
</html>