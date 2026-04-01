export async function withDelay<T>(value: T, timeout = 120): Promise<T> {
  await new Promise((resolve) => window.setTimeout(resolve, timeout))
  return value
}
