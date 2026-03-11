import express from "express";
import { getSnapshot } from "./registry.js";

export const metricsRouter = express.Router();

metricsRouter.get("/metrics", (_req, res) => {
  res.json(getSnapshot());
});

metricsRouter.get("/health", (_req, res) => {
  res.json({ status: "ok", ts: new Date().toISOString() });
});
