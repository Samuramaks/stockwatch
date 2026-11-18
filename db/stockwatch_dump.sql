--
-- PostgreSQL database dump
--

\restrict Urg8HhPJzfrw21StIVG0gMeTF8B0XlBHY10j0WLfjL2gNr0gt3QTBmXDItxWEMS

-- Dumped from database version 18.6
-- Dumped by pg_dump version 18.7

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: printer_statuses; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.printer_statuses (
    printer_id bigint NOT NULL,
    level_pct integer NOT NULL,
    last_polled_at timestamp with time zone NOT NULL,
    last_alerted_at timestamp with time zone
);


ALTER TABLE public.printer_statuses OWNER TO postgres;

--
-- Name: printers; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.printers (
    id bigint NOT NULL,
    name text NOT NULL,
    ip text NOT NULL,
    model text NOT NULL,
    location text NOT NULL,
    snmp_community text NOT NULL,
    cartridge_sku text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.printers OWNER TO postgres;

--
-- Name: printers_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.printers_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.printers_id_seq OWNER TO postgres;

--
-- Name: printers_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.printers_id_seq OWNED BY public.printers.id;


--
-- Name: stock_items; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.stock_items (
    id bigint NOT NULL,
    name text NOT NULL,
    sku text NOT NULL,
    quantity integer DEFAULT 0 NOT NULL,
    threshold integer DEFAULT 0 NOT NULL,
    unit text DEFAULT 'шт'::text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.stock_items OWNER TO postgres;

--
-- Name: stock_items_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.stock_items_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.stock_items_id_seq OWNER TO postgres;

--
-- Name: stock_items_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.stock_items_id_seq OWNED BY public.stock_items.id;


--
-- Name: printers id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.printers ALTER COLUMN id SET DEFAULT nextval('public.printers_id_seq'::regclass);


--
-- Name: stock_items id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_items ALTER COLUMN id SET DEFAULT nextval('public.stock_items_id_seq'::regclass);


--
-- Data for Name: printer_statuses; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.printer_statuses (printer_id, level_pct, last_polled_at, last_alerted_at) FROM stdin;
2	8	2026-10-17 09:18:33.698613+03	\N
3	64	2026-11-17 08:11:52.54748+03	\N
4	8	2026-11-17 08:11:52.54748+03	\N
6	6	2026-11-17 08:11:52.54748+03	\N
\.


--
-- Data for Name: printers; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.printers (id, name, ip, model, location, snmp_community, cartridge_sku, created_at) FROM stdin;
2	Бухгалтерия	192.168.1.10	HP M404	Каб. 12	public	HP26A	2026-10-17 09:17:18.80957+03
3	Отдел кадров	192.0.2.12	ECOSYS M2040dn	Каб. 7	public	TK-1170	2026-11-17 08:23:27.667184+03
4	Бухгалтерия	192.0.2.10	ECOSYS M2040dn	Каб. 12	public	TK-1170	2026-11-17 08:23:27.667184+03
5	Новый принтер	192.0.2.13	ECOSYS M2040dn		public	TK-9999	2026-11-17 08:23:27.667184+03
6	Приёмная	192.0.2.11	ECOSYS M3540dn	Каб. 1	public	TK-3130	2026-11-17 08:23:27.667184+03
\.


--
-- Data for Name: stock_items; Type: TABLE DATA; Schema: public; Owner: postgres
--

COPY public.stock_items (id, name, sku, quantity, threshold, unit, created_at, updated_at) FROM stdin;
2	Бумага A4, 500 л.	PAPER-A4	40	10	пачка	2026-11-17 08:23:41.667205+03	2026-11-17 08:23:41.667205+03
3	Картридж TK-5140	TK-5140	7	3	шт	2026-11-17 08:23:41.667205+03	2026-11-17 11:28:48.08062+03
5	Картридж TK-1170	TK-1170	21	3	шт	2026-11-17 08:23:41.667205+03	2026-11-17 16:16:30.411311+03
4	Картридж TK-3130	TK-3130	22	2	шт	2026-11-17 08:23:41.667205+03	2026-11-17 16:34:31.293132+03
\.


--
-- Name: printers_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.printers_id_seq', 6, true);


--
-- Name: stock_items_id_seq; Type: SEQUENCE SET; Schema: public; Owner: postgres
--

SELECT pg_catalog.setval('public.stock_items_id_seq', 5, true);


--
-- Name: printer_statuses printer_statuses_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.printer_statuses
    ADD CONSTRAINT printer_statuses_pkey PRIMARY KEY (printer_id);


--
-- Name: printers printers_ip_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.printers
    ADD CONSTRAINT printers_ip_key UNIQUE (ip);


--
-- Name: printers printers_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.printers
    ADD CONSTRAINT printers_pkey PRIMARY KEY (id);


--
-- Name: stock_items stock_items_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_items
    ADD CONSTRAINT stock_items_pkey PRIMARY KEY (id);


--
-- Name: stock_items stock_items_sku_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.stock_items
    ADD CONSTRAINT stock_items_sku_key UNIQUE (sku);


--
-- Name: printer_statuses printer_statuses_printer_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.printer_statuses
    ADD CONSTRAINT printer_statuses_printer_id_fkey FOREIGN KEY (printer_id) REFERENCES public.printers(id) ON DELETE CASCADE;


--
-- PostgreSQL database dump complete
--

\unrestrict Urg8HhPJzfrw21StIVG0gMeTF8B0XlBHY10j0WLfjL2gNr0gt3QTBmXDItxWEMS

