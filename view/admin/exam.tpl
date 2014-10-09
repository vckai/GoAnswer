<link href="/public/stylesheets/main.css" rel="stylesheet">
<script src="/public/javascripts/jquery-1.7.2.min.js"></script>

<div>
    <h1 id="title">Add Exam</h1>
    <div class="hr"></div>
    <form id="form1" name="form1" action="" method="post">
        <div class="col-240" id="paramTitle">
            <h4>Question</h4>
            <h4>Options</h4>
            <h4></h4>
            <h4></h4>
            <h4></h4>
            <h4>Answer</h4>
            <h4>Resolve</h4>
        </div>
        <div class="col-400" id="paramInput">
            <input type="text" name="question" value="" />
            <input type="text" name="options[]" value="" />
            <input type="text" name="options[]" value="" />
            <input type="text" name="options[]" value="" />
            <input type="text" name="options[]" value="" />
            <input type="text" name="answer" value="" />
            <input type="text" name="resolve" value="" />
        </div>
        <a href="javascript:;" class="submit button-grey">Add<span></span></a>
    </form>
</div>

<script type="text/javascript">
    var isRept = false;
    $(".submit").click(function(e) {
        if( isRept == true) return false;

        isRept = true;
        $.post("/addExam", $("#form1").serialize(), function(data) {
            isRept = false;

            if(data.suc == 1) {
                alert('添加成功');
                window.location.reload();
            } else if( data.msg ) {
                alert(data.msg);
            } else {
                alert('添加失败');
            }
        });
    });
</script>