/**
 * 公共路由别名
 # 存放系统级公共路由路径，如布局容器、登录页等   
 */
export enum RoutesAlias {
  Layout = '/index/index', // 布局容器
  Login = '/account/auth/login', // 登录页
  AuthCallback = '/account/auth/callback', // centralized_login 回调页
  SocialCallback = '/account/auth/social-callback' // social oauth 中转页
}
