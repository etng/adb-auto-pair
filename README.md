# Adb Auto Pair
运行命令后用手机扫码即可配对远程调试

# 原理

参考自[[Android] Using ADB QR code pairing in R](https://medium.com/@shakalaca/android-using-adb-qr-code-pairing-in-r-52f16db3df6d)

* 无限调试确实方便，少了一根数据线
* 不想打开 AS 要自己连接挨个敲命令我是累了，研究了一下扫码配对的方法
* QR code 格式：`WIFI:T:ADB;S:<name>;P:<code>;;`
    * name 为自定义的 instance name 随机一个就好
    * code 为配对码，就当psk好了
    * 这两个我都用随机生成字符串
* 等手机上开了“使用二维码配对设备”扫描之后，这条dns记录就在局域网里面有了
* 假设 name=debug, code=13456则扫出来的数据是这样子的`WIFI:T:ADB;S:debug;P:123456;;`
* 执行`dns-sd -L debug _adb-tls-pairing._tcp`查询局域网里面广播的dns，结果如下:
* `debug._adb-tls-pairing._tcp.local. can be reached at Android.local.:38362 (interface 4)`
* `Android.local.:38362`就是手机的地址了
* 执行 `adb pair Android.local.:38362 123456` 连连看
* `adb devices -l`就会有自己的手机了
* 比起下面这种一个个敲感觉要方便一些`adb pair 192.168.86.26:41776 563475`
* 生成什么的都要自己整就麻烦，但是我们可以写代码嘛，今天就撸了一个，只能说`it works`,后面再来调整吧
