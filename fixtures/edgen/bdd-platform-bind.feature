# language: zh-CN
Feature: 社交平台绑定状态
  作为 Edgen 用户，我需要在设置中查看各社交平台是否已绑定。

  Background:
    Given 用户已使用有效 token 登录 Edgen

  Scenario: 查看 Twitter 绑定状态（已绑定）
    When 客户端请求查询 Twitter 平台绑定
    Then 应返回 HTTP 200
    And 业务码 code 应为 0
    And data.platform 应为 TWITTER
    And data.bound 应为 true

  Scenario: 查看 Twitter 绑定状态（未绑定）
    When 客户端请求查询 Twitter 平台绑定
    Then 应返回 HTTP 200
    And 业务码 code 应为 0
    And data.bound 应为 false

  Scenario: 分页查看全部平台绑定列表
    When 客户端请求平台绑定列表第 1 页
    Then 应返回 HTTP 200
    And 业务码 code 应为 0
    And data.rows 不应为空

  Scenario: 无效 token 无法查询绑定
    Given 用户使用过期 token
    When 客户端请求查询 Twitter 平台绑定
    Then 应返回 HTTP 401
