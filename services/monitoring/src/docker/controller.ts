import { execFile } from "node:child_process";
import { promisify } from "node:util";
import { MONITORING_ROOT } from "../config/env.js";
import { logger } from "../logger/logger.js";
import { AsyncMutex } from "./lock.js";
import { withRetry } from "../orchestration/backoff.js";

const execFileAsync = promisify(execFile);

const dockerMutex = new AsyncMutex();

interface ExecResult {
  stdout: string;
  stderr: string;
}

async function runDockerCompose(
  composePath: string,
  args: string[]
): Promise<ExecResult> {
  const fullArgs = ["-f", composePath, ...args];
  logger.debug({ composePath, args }, "docker compose");
  const result = await execFileAsync("docker", ["compose", ...fullArgs], {
    cwd: MONITORING_ROOT,
    env: process.env,
  });
  return result;
}

export async function ensureServiceUp(
  composePath: string,
  service: string
): Promise<void> {
  await ensureServicesUp(composePath, [service]);
}

/** Start multiple services from the same compose file in a single docker call. */
export async function ensureServicesUp(
  composePath: string,
  serviceNames: string[]
): Promise<void> {
  await dockerMutex.run(() =>
    withRetry(
      async () => {
        await runDockerCompose(composePath, ["up", "-d", ...serviceNames]);
        logger.info({ composePath, services: serviceNames }, "services ensured up");
      },
      {
        maxAttempts: 4,
        baseDelayMs: 1000,
        label: `ensureServicesUp(${serviceNames.join(",")})`,
      }
    )
  );
}

/** Scale a service to an exact replica count via --scale. */
export async function scaleService(
  composePath: string,
  service: string,
  count: number
): Promise<void> {
  await dockerMutex.run(() =>
    withRetry(
      async () => {
        await runDockerCompose(composePath, [
          "up", "-d", "--scale", `${service}=${count}`,
        ]);
        logger.info({ composePath, service, count }, "service scaled");
      },
      {
        maxAttempts: 4,
        baseDelayMs: 1000,
        label: `scaleService(${service}=${count})`,
      }
    )
  );
}

export async function runOneOffJob(
  composePath: string,
  service: string,
  extraArgs: string[] = []
): Promise<string> {
  return dockerMutex.run(() =>
    withRetry(
      async () => {
        const result = await runDockerCompose(composePath, [
          "run",
          "-d",
          "--rm",
          service,
          ...extraArgs,
        ]);
        const containerId = result.stdout.trim();
        logger.info(
          { composePath, service, containerId },
          "one-off job started"
        );
        return containerId;
      },
      { maxAttempts: 4, baseDelayMs: 1000, label: `runOneOffJob(${service})` }
    )
  );
}

interface DockerPsEntry {
  Name: string;
  State: string;
  Status: string;
  [key: string]: unknown;
}

export async function listRunningContainers(): Promise<DockerPsEntry[]> {
  const result = await execFileAsync(
    "docker",
    ["ps", "--format", "{{json .}}"],
    { cwd: MONITORING_ROOT }
  );
  return result.stdout
    .split("\n")
    .filter(Boolean)
    .map((line) => JSON.parse(line) as DockerPsEntry);
}

export async function isContainerRunning(containerName: string): Promise<boolean> {
  const containers = await listRunningContainers();
  return containers.some(
    (c) =>
      (c["Names"] as string | undefined)
        ?.split(",")
        .some((n) => n.trim().replace(/^\//, "") === containerName) ?? false
  );
}

export async function countRunningJobsByImage(
  composePath: string,
  service: string
): Promise<number> {
  try {
    const result = await execFileAsync(
      "docker",
      ["compose", "-f", composePath, "ps", "--format", "json"],
      { cwd: MONITORING_ROOT }
    );
    const entries = result.stdout
      .split("\n")
      .filter(Boolean)
      .map((l) => {
        try {
          return JSON.parse(l) as DockerPsEntry;
        } catch {
          return null;
        }
      })
      .filter(Boolean) as DockerPsEntry[];

    return entries.filter(
      (e) =>
        e["Service"] === service &&
        (e["State"] === "running" || e["Status"]?.includes("Up"))
    ).length;
  } catch {
    return 0;
  }
}
