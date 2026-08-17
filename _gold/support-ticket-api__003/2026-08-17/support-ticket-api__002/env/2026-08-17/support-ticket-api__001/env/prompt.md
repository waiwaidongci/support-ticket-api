# 项目：小型工单系统API

从0做一个Go小型工单系统API，用Gin和PostgreSQL实现，面向客服场景。用户角色包含客户、客服和主管：客户可以提交工单并查看处理进度；客服可以领取工单、更新状态、填写处理结果和备注；主管可以分配工单、设置优先级、查看统计。工单包含类型、优先级、状态、创建时间、处理人和完整状态变更历史，支持按状态、优先级、负责人筛选，超过SLA时限的工单要突出显示。代码按Go企业项目结构组织：cmd/api/main.go、cmd/migrate、internal/config、internal/model、internal/repository、internal/service、internal/handler、internal/middleware、internal/router、internal/pkg、migrations。状态流转规则放在service层，状态变更必须记录历史，工单分配使用事务保证不会重复分配。
