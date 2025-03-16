"use client";

import React from "react";
import { usePathname, useRouter } from "next/navigation";
import { CloudServerOutlined, CreditCardOutlined, ProfileOutlined, SettingOutlined } from "@ant-design/icons";
import { ConfigProvider, Layout, Menu, MenuProps, theme } from "antd";
import MenuItem from "antd/es/menu/MenuItem";
import trTr from "antd/locale/tr_TR";

const { Header, Sider } = Layout;

interface ExtendedMenuProps extends MenuProps {
  path?: string;
}

type MenuItem = Required<ExtendedMenuProps>["items"][number];

const RootLayout = ({ children }: React.PropsWithChildren) => {
  const {
    token: { colorBgContainer },
  } = theme.useToken();

  const router = useRouter();
  const pathname = usePathname();

  function getItem(label: React.ReactNode, key: React.Key, path: string, icon?: React.ReactNode): MenuItem {
    return {
      key,
      icon,
      label,
      path,
      onClick: () => router.push(path),
    } as MenuItem;
  }

  const menuItems = [
    getItem("Faturalar", "1", "/", <ProfileOutlined />),
    getItem("Ödeme Yöntemleri", "2", "/payment-methods", <CreditCardOutlined />),
    getItem("Hizmetler", "3", "/services", <CloudServerOutlined />),
    getItem("Ayarlar", "4", "/settings", <SettingOutlined />),
  ] as MenuItem[];

  const selectedItemKey =
    menuItems
      .find((menuItem) => {
        const item = menuItem as ExtendedMenuProps;
        if (!item?.path || !pathname) return false;
        return pathname === item.path || pathname.startsWith(item.path + "/");
      })
      ?.key?.toString() ?? "1";

  return (
    <html lang="en">
      <body>
        <ConfigProvider
          theme={{
            algorithm: theme.defaultAlgorithm,
            token: { borderRadius: 0 },
          }}
          locale={trTr}
        >
          <Layout hasSider>
            <Sider
              style={{ overflow: "auto", minHeight: "100vh", insetInlineStart: 0, top: 0, bottom: 0 }}
              theme="light"
            >
              <Menu theme="light" mode="inline" items={menuItems} selectedKeys={[selectedItemKey]} />
            </Sider>
            <Layout>
              <Header style={{ background: colorBgContainer, fontSize: "24px" }}>Faturalar</Header>
              {children}
            </Layout>
          </Layout>
        </ConfigProvider>
      </body>
    </html>
  );
};

export default RootLayout;
