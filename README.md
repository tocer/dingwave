# 为什么会有这个项目？

在实战红蓝对抗中，红队通常通过钓鱼、EDR 利用等手段获得办公网内员工机器的控制权，**后续需要借助后渗透工具从 IM 应用中提取敏感信息进行更深入的信息收集。**

# 使用方法

支持输入钉钉数据库源文件和加密文件，如果是加密文件则需要指定解密密钥，目前仅实现了针对 v2 版本的解密。

其解密密钥为用户 UID，可以查看目录结构了解用户 UID: `C:\Users\{用户名}\AppData\Roaming\DingTalk\{uid}_{version}`。

## 快速使用

输入为解密后的数据库文件
```
./dingwave -d dingtalk.db
```

输入为加密的数据库文件

```
./dingwave -d dingtalk_encrypt.db -k 666165872
```

希望保存解密后的数据库文件

```
./dingwave -d dingtalk_encrypt.db -k 666165872 -o dingtalk.db
```

# 已实现功能

## 查看所有会话

参考了钉钉GUI的设计，将会话区分为置顶、单聊与群聊，按最后一条消息发送日期降序展示：

<img width="2784" height="1602" alt="802b458c973f48027c82856d4b317d28" src="https://github.com/user-attachments/assets/93a71ef3-b2eb-4ffa-a5c7-4fe41e91161e" />

点击某个具体的会话类型后，会展开同类型的所有会话：

![d859893f448fed9f9b350cd677dda591](https://github.com/user-attachments/assets/7d79d8d9-0ae8-4607-ab4f-a9d5d7b42c07)

## 完整解析钉钉消息数据类型

程序已经实现了钉钉内大部分数据类型进行解析，无奈有的数据（例如附件、头像）是存储在本地的，仅通过数据库无法获取更多有效信息，**因此暂时只提供了部分图片以及部分附件的解析。**

<img width="2784" height="1602" alt="73218ec4fdfe65ba0e69461565b8c44e" src="https://github.com/user-attachments/assets/3b355698-df80-4f95-87ca-f0e60701ad8d" />

## 全局搜索聊天记录

<img width="2784" height="1602" alt="a74a24b705e9b8f420edfdb90e54ccc1" src="https://github.com/user-attachments/assets/f12158c3-7f76-4198-99f5-117aa6ec31e7" />

点击某个搜索会话，会展开所有搜索到的消息，同时高亮搜索内容：

<img width="522" height="944" alt="794eea21a999ced13bf7a9d9840ce93b" src="https://github.com/user-attachments/assets/0be637ac-e028-45ff-b494-f838884fff2d" />

点击会话内的具体搜索结果，会跳转到对应的聊天片段中，可以像正常查阅聊天内容一样上滑与下拉：

![695033e64497021be2460ed2fb6b160c](https://github.com/user-attachments/assets/335e21c6-3db9-4cf1-8992-a8ead93936a5)

**当然也支持指定某个会话进行搜索，实现效果与全局搜索效果一致，这里不再赘述。**

## 查看联系人列表

![4d5c7488800aaf866eebdd52518185a5](https://github.com/user-attachments/assets/669426fe-3d60-48aa-a270-31b24878365f)
