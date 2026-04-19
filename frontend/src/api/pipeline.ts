import client from './client'
import type {
  PipelineBootstrap,
  PipelineJourney,
  PipelineLogsResponse,
} from '@/types/pipeline'

export interface BootstrapOpts {
  window?: string
  tickerLimit?: number
}

export async function fetchPipelineBootstrap(
  appId: string,
  opts: BootstrapOpts = {},
): Promise<PipelineBootstrap> {
  const { data } = await client.get<PipelineBootstrap>(
    `/apps/${appId}/pipeline/bootstrap`,
    {
      params: {
        window: opts.window ?? '1h',
        tickerLimit: opts.tickerLimit ?? 50,
      },
    },
  )
  return data
}

export async function fetchLogJourney(
  appId: string,
  logId: string,
): Promise<PipelineJourney> {
  const { data } = await client.get<PipelineJourney>(
    `/apps/${appId}/pipeline/logs/${logId}/journey`,
  )
  return data
}

export interface PipelineLogsQuery {
  since?: string
  until?: string
  limit?: number
  offset?: number
}

export async function fetchPipelineLogs(
  appId: string,
  q: PipelineLogsQuery = {},
): Promise<PipelineLogsResponse> {
  const { data } = await client.get<PipelineLogsResponse>(
    `/apps/${appId}/pipeline/logs`,
    {
      params: {
        since: q.since,
        until: q.until,
        limit: q.limit ?? 50,
        offset: q.offset ?? 0,
      },
    },
  )
  return data
}
