"use client";

import { useEffect, useState } from "react";
import { DownloadOutlined } from "@ant-design/icons";
import { Alert, Button, Flex, Input, Layout, Table, Tag } from "antd";
import type { TableProps } from "antd";

const { Search } = Input;
const { Content } = Layout;

type Invoice = {
  id: string;
  serviceName: "DMP" | "SSP";
  amount: number;
  date: string;
  status: "PAID" | "UNPAID" | "PENDING";
};

const columns: TableProps<Invoice>["columns"] = [
  {
    title: "Servis Adı",
    dataIndex: "serviceName",
    key: "serviceName",
    sorter: (a, b) => a.serviceName.localeCompare(b.serviceName),
  },
  {
    title: "Fatura Numarası",
    dataIndex: "id",
    key: "id",
    sorter: (a, b) => a.id.localeCompare(b.id),
  },
  {
    title: "Tarih",
    dataIndex: "date",
    key: "date",
    render: (date) => (
      <p>{new Date(date).toLocaleDateString("en-US", { year: "numeric", month: "long", day: "numeric" })}</p>
    ),
    sorter: (a, b) => {
      const dateA = new Date(a.date);
      const dateB = new Date(b.date);
      return dateA.getTime() - dateB.getTime();
    },
  },
  {
    title: "Tutar",
    dataIndex: "amount",
    key: "amount",
    sorter: (a, b) => a.amount - b.amount,
  },
  {
    title: "Durum",
    dataIndex: "status",
    key: "status",
    sorter: (a, b) => a.status.localeCompare(b.status),
    render: (text: Invoice["status"]) => {
      if (text === "PAID") return <Tag color="green">Ödendi</Tag>;
      if (text === "UNPAID") return <Tag color="red">Ödenmedi</Tag>;
      if (text === "PENDING") return <Tag color="orange">Bekliyor</Tag>;
    },
  },
  {
    key: "action",
    render: () => <a>Göster</a>,
  },
];

const Home = () => {
  const [invoices, setInvoices] = useState<Array<Invoice>>([]);
  const [isLoading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<Error | undefined>(undefined);

  function onSearch(value: string) {
    setLoading(true);
    fetch(process.env.NEXT_PUBLIC_API_URL + "/invoices?search=" + value)
      .then((res) => res.json())
      .then((data) => setInvoices(data))
      .catch((error) => setError(error))
      .finally(() => setLoading(false));
  }

  useEffect(() => {
    fetch(process.env.NEXT_PUBLIC_API_URL + "/invoices")
      .then((res) => res.json())
      .then((data) => setInvoices(data))
      .catch((error) => setError(error))
      .finally(() => setLoading(false));
  }, []);

  if (error) return <Alert message="Error" description={error.message} type="error" />;

  return (
    <Content>
      <Flex justify="space-between" style={{ margin: 50 }}>
        <Search placeholder="Fatura ara" style={{ width: 320 }} onSearch={onSearch} />
        <Button icon={<DownloadOutlined />} />
      </Flex>
      <Table<Invoice> columns={columns} dataSource={invoices} style={{ margin: 50 }} loading={isLoading} />
    </Content>
  );
};

export default Home;
