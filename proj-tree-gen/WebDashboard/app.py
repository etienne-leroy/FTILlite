"""
FinTracer Web Dashboard
Real-time monitoring and control interface
"""

from flask import Flask, render_template, jsonify, request, session
from flask_socketio import SocketIO, emit
import json
import os
import sys
from datetime import datetime, timedelta
import threading
import time

app = Flask(__name__)
app.config['SECRET_KEY'] = 'fintracer-secret-key'
socketio = SocketIO(app, cors_allowed_origins="*")

class FinTracerDashboard:
    def __init__(self):
        self.active_analyses = {}
        self.system_metrics = {
            'cpu_usage': 0,
            'memory_usage': 0,
            'network_activity': 0,
            'peer_status': {}
        }
        self.analysis_queue = []
        
    def get_system_status(self):
        """Get current system status"""
        return {
            'status': 'operational',
            'active_analyses': len(self.active_analyses),
            'queued_analyses': len(self.analysis_queue),
            'peer_nodes': 4,
            'uptime': '2 days, 14 hours',
            'last_update': datetime.now().isoformat()
        }
    
    def get_recent_results(self, limit=10):
        """Get recent analysis results"""
        # Placeholder data - would come from actual results storage
        return [
            {
                'id': f'analysis_{i}',
                'model': ['Linear', 'Non-Linear', 'Tree', 'Cyclic'][i % 4],
                'timestamp': (datetime.now() - timedelta(hours=i)).isoformat(),
                'accounts_found': [15, 23, 8, 31][i % 4],
                'risk_level': ['HIGH', 'MEDIUM', 'LOW', 'HIGH'][i % 4],
                'execution_time': [12.5, 23.1, 8.9, 45.2][i % 4]
            } for i in range(limit)
        ]

dashboard = FinTracerDashboard()

@app.route('/')
def index():
    """Main dashboard page"""
    return render_template('dashboard.html')

@app.route('/api/status')
def api_status():
    """System status API endpoint"""
    return jsonify(dashboard.get_system_status())

@app.route('/api/results')
def api_results():
    """Recent results API endpoint"""
    limit = request.args.get('limit', 10, type=int)
    return jsonify(dashboard.get_recent_results(limit))

@app.route('/api/start_analysis', methods=['POST'])
def api_start_analysis():
    """Start new analysis API endpoint"""
    data = request.get_json()
    model_type = data.get('model_type')
    parameters = data.get('parameters', {})
    
    # Queue analysis for execution
    analysis_id = f"analysis_{len(dashboard.analysis_queue) + 1}"
    dashboard.analysis_queue.append({
        'id': analysis_id,
        'model_type': model_type,
        'parameters': parameters,
        'status': 'queued',
        'created': datetime.now().isoformat()
    })
    
    return jsonify({'analysis_id': analysis_id, 'status': 'queued'})

@socketio.on('connect')
def handle_connect():
    """Handle client connection"""
    emit('status_update', dashboard.get_system_status())

@socketio.on('request_update')
def handle_update_request():
    """Handle client update request"""
    emit('status_update', dashboard.get_system_status())
    emit('results_update', dashboard.get_recent_results())

def background_metrics_update():
    """Background thread for updating metrics"""
    while True:
        # Update system metrics
        dashboard.system_metrics['cpu_usage'] = 45.2  # Placeholder
        dashboard.system_metrics['memory_usage'] = 62.8  # Placeholder
        
        # Emit updates to connected clients
        socketio.emit('metrics_update', dashboard.system_metrics)
        time.sleep(5)

if __name__ == '__main__':
    # Start background metrics thread
    metrics_thread = threading.Thread(target=background_metrics_update)
    metrics_thread.daemon = True
    metrics_thread.start()
    
    socketio.run(app, debug=True, host='0.0.0.0', port=5000)
