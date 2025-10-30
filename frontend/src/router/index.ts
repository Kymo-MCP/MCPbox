import { createRouter, createWebHistory } from 'vue-router'
import Layout from '@/layouts/index.vue'
import NProgress from '@/utils/nprogress'
import { useUserStore } from '@/stores'

const router = createRouter({
  history: createWebHistory('./'),
  routes: [
    {
      path: '/',
      component: Layout,
      redirect: '/home',
      children: [
        {
          path: '/home',
          name: 'home',
          meta: {
            title: 'home', // title与国际化数据字段对应为了方便国际化处理
            affix: true,
            keepAlive: true,
            isMenu: true,
          },
          component: () => import('../pages/home/index.vue'),
        },
        {
          path: '/template-manage',
          name: 'templateManage',
          meta: {
            title: 'templateManage',
            isMenu: true,
          },
          component: () => import('../pages/mcp/template-manage/index.vue'),
        },
        {
          path: '/new-template',
          name: 'newTemplate',
          meta: {
            title: 'newTemplate',
          },
          component: () => import('../pages/mcp/template-manage/form-template.vue'),
        },
        {
          path: '/instance-manage',
          name: 'instanceManage',
          meta: {
            title: 'instanceManage',
            isMenu: true,
          },
          component: () => import('../pages/mcp/instance-manage/index.vue'),
        },
        {
          path: '/new-instance',
          name: 'newInstance',
          meta: {
            title: 'newInstance',
          },
          component: () => import('../pages/mcp/instance-manage/form-instance.vue'),
        },
        {
          path: '/working-environment',
          name: 'workingEnvironment',
          meta: {
            title: 'runEnviroment',
            isMenu: true,
          },
          component: () => import('../pages/environment/working-environment/index.vue'),
        },
        {
          path: '/pvc-manage',
          name: 'pvcManage',
          meta: {
            title: 'pvcManage',
          },
          component: () => import('../pages/environment/PVC-manage/index.vue'),
        },
        {
          path: '/node-manage',
          name: 'nodeManage',
          meta: {
            title: 'nodeManage',
          },
          component: () => import('../pages/environment/node-manage/index.vue'),
        },
        {
          path: '/code-list',
          name: 'codeList',
          meta: {
            title: 'codeList',
            isMenu: true,
          },
          component: () => import('../pages/script-code/code-lis/index.vue'),
        },
        {
          path: '/update-code-package',
          name: 'updateCodePackage',
          meta: {
            title: 'updateCodePackage',
          },
          component: () => import('../pages/script-code/update/index.vue'),
        },
        {
          path: '/view-code-package',
          name: 'viewCodePackage',
          meta: {
            title: 'viewCodePackage',
          },
          component: () => import('../pages/script-code/view/index.vue'),
        },
        {
          path: '/debug-tools',
          name: 'debugTools',
          meta: {
            title: 'debugTools',
          },
          component: () => import('../pages/debug-tools/index.vue'),
        },
        {
          path: '/user-center',
          name: 'userCenter',
          component: () => import('../pages/login/user-center.vue'),
        },
      ],
    },
    {
      path: '/instance-log',
      name: 'instanceLog',
      component: () => import('../pages/mcp/instance-manage/instance-log.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../pages/login/index.vue'),
    },
    {
      path: '/404',
      name: '404',
      component: () => import('../pages/error/404.vue'),
    },
  ],
})

const whiteList = ['/login'] // don't need login status

router.beforeEach(async (to, from, next) => {
  NProgress.start()

  try {
    // expose login status for global
    const isLogin = useUserStore().isLogin()

    // Has not login
    if (!isLogin) {
      if (whiteList.includes(to.path)) {
        next()
      } else {
        next(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
        NProgress.done()
      }
      return
    }

    // already login and jump to login page lead to home page
    if (to.path === '/login') {
      next({ path: '/' })
      return
    }

    // don't have the path
    if (to.matched.length === 0) {
      next('/404')
      return
    }

    // Handle set page title
    const title = (to.params.title as string) || (to.query.title as string)
    if (title) {
      to.meta.title = title
    }

    next()
  } catch (error) {
    console.error('❌ Route guard error:', error)
    try {
    } catch (resetError) {
      console.error('❌ Failed to reset user state:', resetError)
    }
    next('/login')
    NProgress.done()
  }
})

router.afterEach(() => {
  NProgress.done()
})

export default router
