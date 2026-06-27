# device-and-queues: Instance, Device, and Queue Families

## Guideline

Create a `VkInstance`, select a `VkPhysicalDevice`, create a logical `VkDevice` with the queues you need, and pick queue families by capability — a graphics family, and where the hardware exposes them a dedicated async-compute and a dedicated transfer family — requesting only the features/extensions you use.

## Rationale

Vulkan has no implicit context: every object hangs off a `VkDevice`, which is created from a chosen `VkPhysicalDevice`, which is enumerated from the `VkInstance`. Queue families are how the hardware advertises which engines can run which work; a family with `VK_QUEUE_GRAPHICS_BIT` can do everything, but a dedicated `VK_QUEUE_TRANSFER_BIT`-only family runs DMA copies on a separate engine that overlaps graphics, and a compute-only family enables async compute that overlaps the graphics pipeline. You must request queues at device-creation time (you cannot add them later), and you must enable features (`VkPhysicalDeviceFeatures2` chain) and extensions explicitly — the architecture reason for owning all of this is in gpu-rendering-guide.

## How to Apply

1. Create the `VkInstance` with required instance extensions (surface, plus debug-utils in development) and validation layers when developing.
2. Enumerate physical devices (`vkEnumeratePhysicalDevices`); pick one that supports your surface and required features.
3. Query queue families (`vkGetPhysicalDeviceQueueFamilyProperties`); choose a graphics family that also supports present (`vkGetPhysicalDeviceSurfaceSupportKHR`), and dedicated compute/transfer families when present.
4. Create the `VkDevice` with `VkDeviceQueueCreateInfo` per family, chaining the `VkPhysicalDeviceFeatures2` / `*Vulkan13Features` you enable; retrieve queues with `vkGetDeviceQueue`.

## Example

```c
uint32_t n; vkGetPhysicalDeviceQueueFamilyProperties(phys, &n, NULL);
VkQueueFamilyProperties props[16]; vkGetPhysicalDeviceQueueFamilyProperties(phys, &n, props);

uint32_t gfx = UINT32_MAX, xfer = UINT32_MAX;
for (uint32_t i = 0; i < n; i++) {
    if (props[i].queueFlags & VK_QUEUE_GRAPHICS_BIT) gfx = i;
    // dedicated transfer = TRANSFER but not GRAPHICS/COMPUTE -> separate DMA engine
    if ((props[i].queueFlags & VK_QUEUE_TRANSFER_BIT) &&
        !(props[i].queueFlags & (VK_QUEUE_GRAPHICS_BIT | VK_QUEUE_COMPUTE_BIT))) xfer = i;
}

float prio = 1.0f;
VkDeviceQueueCreateInfo qci[] = {
    {.sType = VK_STRUCTURE_TYPE_DEVICE_QUEUE_CREATE_INFO, .queueFamilyIndex = gfx,
     .queueCount = 1, .pQueuePriorities = &prio},
};
VkPhysicalDeviceVulkan13Features v13 = {.sType = VK_STRUCTURE_TYPE_PHYSICAL_DEVICE_VULKAN_1_3_FEATURES,
    .dynamicRendering = VK_TRUE, .synchronization2 = VK_TRUE};
VkDeviceCreateInfo dci = {.sType = VK_STRUCTURE_TYPE_DEVICE_CREATE_INFO, .pNext = &v13,
    .queueCreateInfoCount = 1, .pQueueCreateInfos = qci,
    .enabledExtensionCount = ext_count, .ppEnabledExtensionNames = exts};
VkDevice dev; vkCreateDevice(phys, &dci, NULL, &dev);
VkQueue gfx_q; vkGetDeviceQueue(dev, gfx, 0, &gfx_q);
```

## Gotchas

- A queue family that reports `GRAPHICS` may be the only one; do not assume a dedicated transfer/compute family exists — fall back to the graphics family.
- Present support is per (family × surface), not a device-wide property — query it with `vkGetPhysicalDeviceSurfaceSupportKHR`, and the present family may differ from graphics.
- Features must be enabled at device creation; using `descriptorIndexing` or `timelineSemaphore` without enabling it is undefined and validation-flagged.
- A resource used on two different queue families needs an ownership transfer, see [references/resources-and-barriers.md](./resources-and-barriers.md).

## Related

[references/resources-and-barriers.md](./resources-and-barriers.md), [references/synchronization.md](./synchronization.md), [references/commands-and-swapchain.md](./commands-and-swapchain.md)
