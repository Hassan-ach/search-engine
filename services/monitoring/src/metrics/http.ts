import express from "express";
import path from "node:path";
import { MONITORING_ROOT } from "../config/env.js";
import { getSnapshot } from "./registry.js";

export const metricsRouter = express.Router();

metricsRouter.use(express.static(path.join(MONITORING_ROOT, "static")));

metricsRouter.get("/metrics", (_req, res) => {
    res.json(getSnapshot());
});

metricsRouter.get("/health", (_req, res) => {
    res.json({ status: "ok", ts: new Date().toISOString() });
});
