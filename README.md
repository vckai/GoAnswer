Go Answer
====================

网页版本的答题王，通过Websocket通讯。

安装：
go get github.com/vckai/GoAnswer

数据是存在Mongodb，可能使用redis会比较好，但是之前用mongodb比较少，也算是一种学习吧。

安装好之后，通过http://127.0.0.1:9999/addExam 添加题库，这个属于比较简单地添加。



至于说明文档，等空了在完善下！
====================


项目功能说明
=========================================================================================================

表说明
---------------------------------------------------------------------------------------------------------
DB_SMS.SmsSendList  短信待发送表
DB_SMS.HrSmsBeSend  短信发送队列表
DB_SMS.HrSmsReport  短信回执状态表

C++的短信发送是写到DB_SMS.SmsSendList，该表的触发器对数据进行拆分成单条短信insert到DB_SMS.HrSmsBeSend中。

而python脚本会轮询DB_SMS.HrSmsBeSend表，每次取1000条短信，启动20个线程调用短信网关接口进行短信发送。



文件结构
=========================================================================================================

程序
---------------------------------------------------------------------------------------------------------
--- config.py		短信发送程序配置文件, DB, SMS通道
--- sendSmsScript.py    短信发送python源代码
--- smsServer.py	接收短信网关回执的python源代码

shell脚本
---------------------------------------------------------------------------------------------------------
--- cutLog.sh		按日分割短信发送，以及回执脚本    
--- runSendScript.sh    短信发送运行脚本，负责将打印信息导入到log文件中
--- runServer.sh	接收短信网关回执运行脚本，负责将打印信息导入到log文件中
--- sendScriptDaemon.sh 短信发送程序守护脚本，防止程序僵死或者进程意外死掉
--- smsServerDaemon.sh  接收短信网关回执守护脚本，防止进程意外死掉

crontab
---------------------------------------------------------------------------------------------------------
--- crontab		所有crontab任务信息



crontab说明
=========================================================================================================
		*/1 * * * * /bin/bash /data/hrSms/runSendScript.sh
		*/1 * * * * /bin/bash /data/hrSms/sendScriptDaemon.sh
		*/1 * * * * /bin/bash /data/hrSms/smsServerDaemon.sh
		#每天凌晨04：30对日记进行分割
		30 04 * * * /bin/bash /data/hrSms/catLog.sh



安装使用说明
=========================================================================================================
1. 安装Python，目前在2.6.6版本测试通过
2. 安装python的mysql模块：MySQLdb，参考：http://blog.csdn.net/wklken/article/details/7271019 。
3. 安装python的web框架：web.py，参考：http://webpy.org/install#install 。
4. 将程序解压到/data/hrSms目录中，如果目录不是在/data/hrSms则需要修改shell脚本以及crontab中的路径地址。
5. 修改config.py的配置，如DB信息。
6. 运行linux命令：crontab -e, 将crontab的任务信息全部复制过去。 

Nginx 转发配置
---------------------------------------------------------------------------------------------------------
域名callback.jzs.so是提供给短信网关的回执回调地址，这里需要通过nginx配置转发到smsServer进程上。
smsServer进程默认是使用9002端口，端口配置在：runServer.sh中。

注意：端口需要nginx以及runServer.sh同时更改。

upstream frontends {
        ip_hash;
        server 127.0.0.1:9002;
}
server {
        listen       80;
        server_name  callback.jzs.so;

        location / {
                proxy_pass_header Server;
                proxy_set_header Host $http_host;
                proxy_redirect off;
                proxy_set_header X-Real-IP $remote_addr;
                proxy_set_header X-Scheme $scheme;
                proxy_pass http://frontends;
        }
        access_log  /var/log/weblog/hrSms.access.log;
}


运行
---------------------------------------------------------------------------------------------------------
运行接收回执的程序
./runServer.sh

短信发送的脚本会由crontab直接拉起。
注意目录下会生成sendSms.run锁文件，该锁文件是防止僵尸程序以及程序进程唯一使用的。

重启
---------------------------------------------------------------------------------------------------------
首先使用ps -axu | grep -E 'sendSmsScript|smsServer' 查找出两个进程的相应pid。
然后使用kill pid命令将进程kill掉，然后删除/data/hrSms目录的sendSms.run锁文件。
再执行./runServer.sh即可。

停止
---------------------------------------------------------------------------------------------------------
首先将crontab的任务删除，然后再参考"重启"的kill方式即可。



TODO
=========================================================================================================
1. 短信预警。
2. 重启/停止处理。
