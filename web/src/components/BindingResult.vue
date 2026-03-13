<template>
  <div v-if="show" class="binding-modal-overlay" @click="closeModal">
    <div class="binding-modal" @click.stop>
      <div class="binding-header">
        <h3>Template Binding Successful</h3>
        <button @click="closeModal" class="close-btn">&times;</button>
      </div>

      <div class="binding-content">
        <div class="binding-info">
          <h4>Binding Information</h4>
          <p><strong>Template:</strong> {{ templateName }}</p>
          <p><strong>Binding Name:</strong> {{ bindingName }}</p>
          <p><strong>Kubeconfig Secret:</strong> {{ kubeconfigSecretName }}</p>
        </div>

        <div class="instructions-section">
          <h4>Consumer Cluster Setup</h4>
          <p class="instructions-text">
            The provider-side binding has been created. To connect your consumer cluster,
            download and apply the following files:
          </p>

          <div class="step-group">
            <h5>1. Deploy the konnector agent</h5>
            <p class="step-description">The konnector syncs resources between provider and consumer clusters.</p>
            <div class="download-block">
              <div class="download-actions">
                <button @click="downloadKonnectorManifests" class="download-btn primary">Download konnector.yaml</button>
              </div>
            </div>
            <div class="command-block">
              <code>kubectl apply -f konnector.yaml</code>
              <button @click="copyCommand('kubectl apply -f konnector.yaml')" class="copy-cmd-btn">Copy</button>
            </div>
          </div>

          <div class="step-group">
            <h5>2. Apply the binding setup</h5>
            <p class="step-description">This creates the namespace, kubeconfig secret, and APIServiceBinding on your consumer cluster.</p>
            <div class="download-block">
              <div class="download-actions">
                <button @click="downloadBindingSetup" class="download-btn primary">Download binding-setup.yaml</button>
              </div>
            </div>
            <div class="command-block">
              <code>kubectl apply -f binding-setup.yaml</code>
              <button @click="copyCommand('kubectl apply -f binding-setup.yaml')" class="copy-cmd-btn">Copy</button>
            </div>
          </div>
        </div>

        <div class="alternative-section">
          <details>
            <summary>Advanced: Download individual files</summary>
            <div class="manual-setup">
              <div class="download-block">
                <div class="download-actions">
                  <button @click="downloadKubeconfig" class="download-btn">Download kubeconfig.yaml</button>
                  <button @click="downloadAPIRequests" class="download-btn">Download apiservice-export.yaml</button>
                </div>
              </div>
            </div>
          </details>
        </div>
      </div>

      <div class="binding-footer">
        <button @click="closeModal" class="ok-btn">Close</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { BindingResponse } from '../types/binding'

interface Props {
  show: boolean
  templateName: string
  bindingResponse: BindingResponse
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
}>()

// Use binding name from response, falling back to template name
const bindingName = computed(() => {
  return props.bindingResponse.bindingName || props.templateName
})

// Generate a stable secret name from the binding name
const kubeconfigSecretName = computed(() => {
  const safeName = bindingName.value.toLowerCase().replace(/[^a-z0-9-]/g, '-')
  return `kubeconfig-${safeName}`
})

const closeModal = () => {
  emit('close')
}

const copyCommand = async (command: string) => {
  try {
    await navigator.clipboard.writeText(command)
  } catch (err) {
    console.error('Failed to copy command:', err)
    const textarea = document.createElement('textarea')
    textarea.value = command
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
  }
}

const triggerDownload = (content: string, filename: string, type = 'text/yaml') => {
  const blob = new Blob([content], { type })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

// Decode kubeconfig - it comes as base64-encoded bytes from the Go backend
const decodedKubeconfig = computed(() => {
  try {
    return atob(props.bindingResponse.kubeconfig)
  } catch {
    return props.bindingResponse.kubeconfig
  }
})

// Generate the binding-setup.yaml YAML bundle
const bindingSetupYaml = computed(() => {
  const kubeconfigBase64 = props.bindingResponse.kubeconfig
  const secretName = kubeconfigSecretName.value
  const name = bindingName.value

  return `apiVersion: v1
kind: Namespace
metadata:
  name: kube-bind
---
apiVersion: v1
kind: Secret
metadata:
  name: ${secretName}
  namespace: kube-bind
  labels:
    kubebind.kube-bind.io/provider-kubeconfig: "true"
type: Opaque
data:
  kubeconfig: ${kubeconfigBase64}
---
apiVersion: kube-bind.io/v1alpha2
kind: APIServiceBinding
metadata:
  name: ${name}
spec:
  kubeconfigSecretRef:
    namespace: kube-bind
    name: ${secretName}
    key: kubeconfig
`
})

const downloadBindingSetup = () => {
  triggerDownload(bindingSetupYaml.value, 'binding-setup.yaml')
}

const downloadKonnectorManifests = async () => {
  try {
    const response = await fetch('/api/konnector-manifests')
    if (!response.ok) {
      throw new Error(`HTTP ${response.status}`)
    }
    const yaml = await response.text()
    triggerDownload(yaml, 'konnector.yaml')
  } catch (error) {
    console.error('Failed to fetch konnector manifests:', error)
  }
}

const downloadKubeconfig = () => {
  triggerDownload(decodedKubeconfig.value, 'kubeconfig.yaml')
}

const downloadAPIRequests = () => {
  try {
    const apiRequestsYaml = props.bindingResponse.requests.map(req => {
      if (typeof req === 'string') {
        return req.trim()
      } else {
        return JSON.stringify(req, null, 2)
      }
    }).join('\n---\n')

    triggerDownload(apiRequestsYaml, 'apiservice-export.yaml')
  } catch (error) {
    console.error('Failed to format API requests:', error)
    const json = JSON.stringify(props.bindingResponse.requests, null, 2)
    triggerDownload(json, 'apiservice-export.json', 'application/json')
  }
}
</script>

<style scoped>
.binding-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Roboto', 'Oxygen', 'Ubuntu', 'Cantarell', sans-serif;
}

.binding-modal {
  background: white;
  border-radius: 12px;
  width: 90%;
  max-width: 900px;
  max-height: 85vh;
  overflow: hidden;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.2);
}

.binding-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem 2rem;
  border-bottom: 1px solid #e5e7eb;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
}

.binding-header h3 {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
}

.close-btn {
  background: none;
  border: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: rgba(255, 255, 255, 0.8);
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: white;
}

.binding-content {
  padding: 2rem;
  max-height: 70vh;
  overflow-y: auto;
}

.binding-info {
  margin-bottom: 2rem;
  padding-bottom: 1.5rem;
  border-bottom: 1px solid #e5e7eb;
}

.binding-info h4 {
  margin-bottom: 1rem;
  color: #111827;
  font-size: 1.125rem;
  font-weight: 600;
}

.binding-info p {
  margin: 0.5rem 0;
  color: #6b7280;
  font-size: 0.9rem;
}

.instructions-section {
  margin-bottom: 2rem;
}

.instructions-section h4 {
  margin-bottom: 1rem;
  color: #111827;
  font-size: 1.125rem;
  font-weight: 600;
}

.instructions-text {
  margin-bottom: 1.5rem;
  color: #6b7280;
  line-height: 1.6;
}

.step-group {
  margin-bottom: 2rem;
}

.step-group h5 {
  margin-bottom: 0.5rem;
  color: #374151;
  font-size: 1rem;
  font-weight: 600;
}

.step-description {
  margin-bottom: 0.75rem;
  color: #6b7280;
  font-size: 0.875rem;
}

.download-block {
  background: #f0f9ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  padding: 1rem;
  margin-bottom: 0.75rem;
}

.download-actions {
  display: flex;
  gap: 1rem;
}

.download-btn {
  padding: 0.75rem 1.5rem;
  background: #6b7280;
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s;
}

.download-btn:hover {
  background: #4b5563;
}

.download-btn.primary {
  background: #3b82f6;
}

.download-btn.primary:hover {
  background: #2563eb;
}

.command-block {
  display: flex;
  align-items: center;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  padding: 1rem;
  gap: 1rem;
}

.command-block code {
  flex: 1;
  font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', 'Consolas', monospace;
  font-size: 0.875rem;
  color: #1f2937;
  background: none;
  word-break: break-all;
}

.copy-cmd-btn {
  padding: 0.5rem 1rem;
  background: #3b82f6;
  color: white;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.875rem;
  font-weight: 500;
  transition: background-color 0.2s;
  flex-shrink: 0;
}

.copy-cmd-btn:hover {
  background: #2563eb;
}

.alternative-section {
  margin-top: 2rem;
  padding-top: 1.5rem;
  border-top: 1px solid #e5e7eb;
}

.alternative-section details {
  cursor: pointer;
}

.alternative-section summary {
  font-weight: 600;
  color: #6b7280;
  padding: 0.5rem 0;
  outline: none;
}

.manual-setup {
  padding: 1rem 0;
}

.binding-footer {
  padding: 1.5rem 2rem;
  border-top: 1px solid #e5e7eb;
  background: #f9fafb;
  text-align: right;
}

.ok-btn {
  padding: 0.75rem 1.5rem;
  background: linear-gradient(135deg, #10b981 0%, #059669 100%);
  color: white;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.ok-btn:hover {
  background: linear-gradient(135deg, #059669 0%, #047857 100%);
  transform: translateY(-1px);
}
</style>
