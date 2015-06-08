<!DOCTYPE html>
<html>
<head>
    <title>爱宝宝APP接口在线调试工具</title>   

    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <!-- js基础库文件 -->
    <link href="/public/stylesheets/main.css" rel="stylesheet">
    <script src="/public/javascripts/jquery-1.7.2.min.js"></script>
</head>

<body>
        <div id="layout">
            <div id="cboxOverlay" style="display:none; background-color:#000; opacity: 0; position:fixed; width:100%; height:100%;top:0; left:0; z-index:999; overflow:hidden;"></div>
            
            <div class="page fixed">
                <div id="sidebar">
                    <ul id="navigation">
                        <li id="/addExam">
                            <div><a href="javascript:;" onclick="loadMethodList('/addExam', this);">Add Exam</a></div>
                            <div class="back"></div>
                        </li>
                    </ul>
                </div>

                <div id="loading" style="z-index:1000; top: 150px; left:550px; position:absolute; display: none;"><img src="/public/images/loading.gif" /></div>
                <div id="content" style="display: none;">
                    <iframe id="load_page" scrolling="no" frameborder="0" style="overflow: hidden; display: none; width:100%" src=""></iframe>
                </div>
            </div>
        </div>

        <script>
            function loadMethodList(url, obj) {
                $("#content").show();
                $("#navigation > li").each(function(ret) {
                    $(this).attr('class', $(this).attr("id") == url ? 'active' : '');
                });
                showload();

                $('#load_page').height("500px");
                $("#load_page").attr("src", url);
                $("#load_page").one("load", function() {
                    console.log('heheheh.');
                    hideload();

                    $('#load_page').css({overflow:"hidden"});
                    $('#load_page').show();
                });
            }

            function showload() {
                var top = $("#content").height() / 2;
                $("#loading").css({"top": top});
                $("#cboxOverlay").show();
                $("#loading").show();
            }

            function hideload() {
                $("#cboxOverlay").hide();
                $("#loading").hide();
            }
        </script>
</body>
</html>