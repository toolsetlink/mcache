# 欢迎使用 mcache 内存缓存包

### mcache是一个基于golang-lru开发的缓存包

mcache 增加了缓存过期时间,增加lfu算法,修改了原有arc算法的依赖结构.
后续还会源源不断增加内存算法.

## 特征
根据过期时间懒汉式删除过期数据,也可主动刷新过期缓存

## why? 为什么要用mcache?
因缓存的使用相关需求,牺牲一部分服务器内存,因减少了网络数据交互,直接使用本机内存,可换取比redis,memcache等更快的缓存速度,
可做为更高一层的缓存需要

## what? 用mcache能做什么?
可作为超高频率数据使用的缓存存储机制
  
## how? mcache怎么用?
根据需要的不同缓存淘汰算法,使用对应的调用 方式


## 现已支持内存算法:
### lru
### lfu
### arc
### 2q
### hashlru
### hashlfu

## 性能对比
hashlru 与 lru 性能对比

算法            | 耗时
-------------  | :-------------
lru            | 220.2s
hashlru-2分区   | 267.75s
hashlru-4分区   | 137.36s
hashlru-8分区   | 22.4s
hashlru-16分区  | 23.57s
hashlru-32分区  | 16.84s
hashlru-64分区  | 15.29s

hashlfu 与 lfu 性能对比

算法            | 耗时
-------------  | :-------------
lru            | 220.92s
hashlfu-2分区   | 231.28s
hashlfu-4分区   | 72.74s
hashlfu-8分区   | 20.33s
hashlfu-16分区  | 17.76s
hashlfu-32分区  | 16.93s
hashlfu-64分区  | 16.03s

### hash算法减少耗时原因:
LruCache在高QPS下的耗时增加原因分析：

线程安全的LruCache中有锁的存在。每次读写操作之前都有加锁操作，完成读写操作之后还有解锁操作。 在低QPS下，锁竞争的耗时基本可以忽略；但是在高QPS下，大量的时间消耗在了等待锁的操作上，导致耗时增长。

HashLruCache适应高QPS场景：

针对大量的同步等待操作导致耗时增加的情况，解决方案就是尽量减小临界区。引入Hash机制，对全量数据做分片处理，在原有LruCache的基础上形成HashLruCache，以降低查询耗时。

HashLruCache引入哈希算法，将缓存数据分散到N个LruCache上。查询时也按照相同的哈希算法，先获取数据可能存在的分片，然后再去对应的分片上查询数据。这样可以增加LruCache的读写操作的并行度，减小同步等待的耗时。

## 代码实现:    
    
    len := 10  
    
    // NewLRU 构造一个给定大小的LRU缓存列表
    Cache, _ := m_cache.NewLRU(Len)

    // Add 向缓存添加一个值。如果已经存在,则更新信息
    Cache.Add(1,1,1614306658000)
    Cache.Add(2,2,0) // expirationTime 传0代表无过期时间

    // Get 从缓存中查找一个键的值
    Cache.Get(2)

    更多方法,请查看 interface

## JetBrains操作系统许可证

durl 是根据JetBrains sro授予的免费JetBrains开源许可证与GoLand一起开发的，因此在此我要表示感谢。

[免费申请 jetbrains 全家桶](https://zhuanlan.zhihu.com/p/264139984?utm_source=wechat_session)


## 交流
#### 如果文档中未能覆盖的任何疑问,欢迎您发送邮件到<songangweb@foxmail.com>,我会尽快答复。
#### 您可以在提出使用中需要改进的地方,我会考虑合理性并尽快修改。
#### 如果您发现 bug 请及时提 issue,我会尽快确认并修改。
#### 有劳点一下 star，一个小小的 star 是作者回答问题的动力 🤝
#### 

