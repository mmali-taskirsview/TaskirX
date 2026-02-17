import Link from 'next/link'
import { ArrowRight, BarChart3, Shield, Target, Zap, Users, Globe } from 'lucide-react'

export default function HomePage() {
  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-purple-900 to-gray-900">
      {/* Header */}
      <header className="border-b border-gray-800">
        <div className="container mx-auto px-6 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <div className="w-10 h-10 bg-gradient-to-br from-purple-500 to-purple-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold text-xl">T</span>
              </div>
              <span className="text-2xl font-bold text-white">
                TaskirX
              </span>
              <span className="text-xs bg-purple-900 text-purple-300 px-2 py-1 rounded-full font-semibold">
                v3.0
              </span>
            </div>
            <nav className="flex items-center space-x-6">
              <Link href="/login" className="text-gray-300 hover:text-white transition">
                Login
              </Link>
              <Link 
                href="/login"
                className="px-4 py-2 bg-gradient-to-r from-purple-600 to-purple-700 text-white rounded-lg hover:shadow-lg transition"
              >
                Get Started
              </Link>
            </nav>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="container mx-auto px-6 py-20">
        <div className="text-center max-w-4xl mx-auto">
          <h1 className="text-6xl font-bold mb-6 text-white">
            The Future of <span className="text-purple-400">Programmatic Advertising</span>
          </h1>
          <p className="text-xl text-gray-300 mb-8 leading-relaxed">
            Enterprise-grade ad exchange platform powered by AI. 
            Real-time bidding, fraud detection, and intelligent optimization—all in one place.
          </p>
          
          {/* Launch App Button */}
          <div className="mt-12 flex justify-center">
            <Link 
              href="/login"
              className="group inline-flex items-center justify-center rounded-full bg-gradient-to-r from-purple-600 to-blue-600 px-10 py-5 text-xl font-bold text-white transition-all hover:scale-105 hover:shadow-[0_0_30px_rgba(124,58,237,0.5)] bg-[length:200%_200%] animate-gradient"
            >
              <div className="flex items-center space-x-3">
                <Shield className="h-6 w-6" />
                <span>Launch App</span>
                <ArrowRight className="h-6 w-6 transition-transform group-hover:translate-x-1" />
              </div>
            </Link>
          </div>
          
          <div className="mt-8 text-center">
             <p className="text-gray-500 text-sm">
               Unified Access for Advertisers, Publishers & Admins
             </p>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-4 gap-8 mt-20 max-w-5xl mx-auto">
          {[
            { label: 'Requests/sec', value: '100K+' },
            { label: 'Latency', value: '<10ms' },
            { label: 'Uptime', value: '99.99%' },
            { label: 'Fraud Detection', value: '95%+' },
          ].map((stat) => (
            <div key={stat.label} className="text-center">
              <div className="text-4xl font-bold text-purple-400">
                {stat.value}
              </div>
              <div className="text-sm text-gray-400 mt-2">{stat.label}</div>
            </div>
          ))}
        </div>
      </section>

      {/* Features */}
      <section className="container mx-auto px-6 py-20">
        <h2 className="text-4xl font-bold text-center mb-16">
          Powered by AI & Machine Learning
        </h2>
        <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
          {[
            {
              icon: <Shield className="w-8 h-8" />,
              title: 'Fraud Detection',
              description: 'Random Forest ML models detect fraudulent traffic in real-time with 95%+ accuracy',
              color: 'from-red-500 to-pink-500',
            },
            {
              icon: <Target className="w-8 h-8" />,
              title: 'Smart Matching',
              description: 'TF-IDF and collaborative filtering ensure perfect ad-audience alignment',
              color: 'from-blue-500 to-cyan-500',
            },
            {
              icon: <Zap className="w-8 h-8" />,
              title: 'Bid Optimization',
              description: 'Thompson Sampling maximizes ROI with intelligent budget pacing',
              color: 'from-purple-500 to-pink-500',
            },
            {
              icon: <BarChart3 className="w-8 h-8" />,
              title: 'Real-time Analytics',
              description: 'ClickHouse-powered dashboards with sub-second query performance',
              color: 'from-green-500 to-emerald-500',
            },
          ].map((feature) => (
            <div key={feature.title} className="bg-white p-6 rounded-xl shadow-lg hover:shadow-2xl transition">
              <div className={`w-16 h-16 bg-gradient-to-br ${feature.color} rounded-lg flex items-center justify-center text-white mb-4`}>
                {feature.icon}
              </div>
              <h3 className="text-xl font-bold mb-2">{feature.title}</h3>
              <p className="text-gray-600">{feature.description}</p>
            </div>
          ))}
        </div>
      </section>

      {/* Architecture */}
      <section className="bg-white py-20">
        <div className="container mx-auto px-6">
          <h2 className="text-4xl font-bold text-center mb-16">
            Built for Scale & Performance
          </h2>
          <div className="grid md:grid-cols-3 gap-8 max-w-5xl mx-auto">
            {[
              {
                title: 'NestJS Backend',
                description: 'TypeScript microservices with multi-tenant architecture',
                tech: ['TypeScript', 'PostgreSQL', 'JWT Auth'],
              },
              {
                title: 'Go Bidding Engine',
                description: 'Ultra-fast RTB with sub-10ms response times',
                tech: ['Go 1.21', 'Redis', '100K QPS'],
              },
              {
                title: 'Python AI Agents',
                description: '3 FastAPI services with ML-powered intelligence',
                tech: ['FastAPI', 'scikit-learn', 'NumPy'],
              },
            ].map((service) => (
              <div key={service.title} className="border-2 border-gray-200 p-6 rounded-xl hover:border-blue-600 transition">
                <h3 className="text-xl font-bold mb-2">{service.title}</h3>
                <p className="text-gray-600 mb-4">{service.description}</p>
                <div className="flex flex-wrap gap-2">
                  {service.tech.map((tech) => (
                    <span key={tech} className="px-3 py-1 bg-blue-100 text-blue-700 rounded-full text-sm">
                      {tech}
                    </span>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="container mx-auto px-6 py-20">
        <div className="bg-gradient-to-r from-blue-600 to-purple-600 rounded-2xl p-12 text-center text-white">
          <h2 className="text-4xl font-bold mb-4">
            Ready to Transform Your Ad Operations?
          </h2>
          <p className="text-xl mb-8 opacity-90">
            Join leading advertisers and publishers using TaskirX
          </p>
          <Link 
            href="/dashboard"
            className="inline-flex items-center space-x-2 px-8 py-4 bg-white text-blue-600 rounded-lg hover:shadow-2xl transition text-lg font-semibold"
          >
            <span>Start Free Trial</span>
            <ArrowRight className="w-5 h-5" />
          </Link>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t bg-gray-50 py-12">
        <div className="container mx-auto px-6">
          <div className="flex items-center justify-between">
            <div className="flex items-center space-x-2">
              <div className="w-8 h-8 bg-gradient-to-br from-blue-600 to-purple-600 rounded-lg flex items-center justify-center">
                <span className="text-white font-bold">T</span>
              </div>
              <span className="text-xl font-bold">TaskirX</span>
            </div>
            <div className="text-sm text-gray-600">
              © 2026 TaskirX. Enterprise Ad Exchange Platform.
            </div>
          </div>
        </div>
      </footer>
    </div>
  )
}
