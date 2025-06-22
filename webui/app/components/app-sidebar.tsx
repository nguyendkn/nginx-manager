import * as React from "react"
import {
  IconActivity,
  IconChartBar,
  IconFileText,
  IconWorld,
  IconHelp,
  IconList,
  IconLock,
  IconDeviceDesktop,
  IconSearch,
  IconServer,
  IconSettings,
  IconShield,
  IconTrendingUp,
} from "@tabler/icons-react"

import { NavConfigs } from "~/components/nav-configs"
import { NavDocuments } from "~/components/nav-documents"
import { NavMain } from "~/components/nav-main"
import { NavSecondary } from "~/components/nav-secondary"
import { NavUser } from "~/components/nav-user"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "~/components/ui/sidebar"

const data = {
  user: {
    name: "Admin User",
    email: "admin@nginx-manager.local",
    avatar: "/avatars/admin.jpg",
  },
  navMain: [
    {
      title: "Dashboard",
      url: "/dashboard",
      icon: IconChartBar,
    },
    {
      title: "Proxy Hosts",
      url: "/proxy-hosts",
      icon: IconWorld,
    },
    {
      title: "SSL Certificates",
      url: "/certificates",
      icon: IconShield,
    },
    {
      title: "Access Lists",
      url: "/access-lists",
      icon: IconLock,
    },
    {
      title: "Monitoring",
      url: "/monitoring",
      icon: IconDeviceDesktop,
    },
    {
      title: "Analytics",
      url: "/analytics",
      icon: IconTrendingUp,
    },
  ],
  navConfigs: [
    {
      title: "Nginx Configs",
      icon: IconFileText,
      isActive: true,
      url: "/nginx-configs",
      items: [
        {
          title: "Active Configs",
          url: "/nginx-configs",
        },
        {
          title: "Create New",
          url: "/nginx-configs/new",
        },
      ],
    },
    {
      title: "Templates",
      icon: IconList,
      url: "/nginx-templates",
      items: [
        {
          title: "All Templates",
          url: "/nginx-templates",
        },
        {
          title: "Create Template",
          url: "/nginx-templates/new",
        },
      ],
    },
    {
      title: "System Health",
      icon: IconActivity,
      url: "/system-health",
      items: [
        {
          title: "Service Status",
          url: "/system-health/services",
        },
        {
          title: "Performance",
          url: "/system-health/performance",
        },
      ],
    },
  ],
  navSecondary: [
    {
      title: "Settings",
      url: "/settings",
      icon: IconSettings,
    },
    {
      title: "Help & Support",
      url: "/help",
      icon: IconHelp,
    },
    {
      title: "Search",
      url: "/search",
      icon: IconSearch,
    },
  ],
  tools: [
    {
      name: "System Logs",
      url: "/logs",
      icon: IconFileText,
    },
    {
      name: "Backup & Restore",
      url: "/backup",
      icon: IconServer,
    },
    {
      name: "API Documentation",
      url: "/api-docs",
      icon: IconFileText,
    },
  ],
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:!p-1.5"
            >
              <a href="/dashboard">
                <IconServer className="!size-5" />
                <span className="text-base font-semibold">Nginx Manager</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={data.navMain} />
        <NavConfigs items={data.navConfigs} />
        <NavDocuments items={data.tools} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
    </Sidebar>
  )
}
