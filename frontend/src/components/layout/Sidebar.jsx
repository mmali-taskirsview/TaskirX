import { NavLink, useLocation } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';
import {
  LayoutDashboard,
  Megaphone,
  Users,
  BarChart3,
  Activity,
  FileText,
  Settings,
  ChevronLeft,
  ChevronRight,
  Zap
} from 'lucide-react';

const Sidebar = ({ isOpen, onToggle }) => {
  const { user, isAdmin } = useAuth();
  const location = useLocation();

  const navItems = [
    { path: '/dashboard', icon: LayoutDashboard, label: 'Dashboard' },
    { path: '/campaigns', icon: Megaphone, label: 'Campaigns' },
    { path: '/analytics', icon: BarChart3, label: 'Analytics' },
    { path: '/rtb-monitor', icon: Activity, label: 'RTB Monitor' },
    { path: '/reports', icon: FileText, label: 'Reports' },
    ...(isAdmin ? [{ path: '/users', icon: Users, label: 'Users' }] : []),
    { path: '/settings', icon: Settings, label: 'Settings' },
  ];

  return (
    <aside 
      className={`fixed left-0 top-0 h-full bg-dark-grey text-white transition-all duration-300 z-50 ${
        isOpen ? 'w-64' : 'w-20'
      }`}
    >
      {/* Logo */}
      <div className="h-16 flex items-center justify-between px-4 border-b border-gray-700">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-gradient-to-br from-cyber-blue to-neon-green rounded-lg flex items-center justify-center">
            <Zap className="w-6 h-6 text-white" />
          </div>
          {isOpen && (
            <span className="font-bold text-xl bg-gradient-to-r from-cyber-blue to-neon-green bg-clip-text text-transparent">
              TaskirX
            </span>
          )}
        </div>
        <button 
          onClick={onToggle}
          className="p-1 rounded-lg hover:bg-gray-700 transition-colors"
        >
          {isOpen ? <ChevronLeft size={20} /> : <ChevronRight size={20} />}
        </button>
      </div>

      {/* Navigation */}
      <nav className="mt-6 px-3">
        {navItems.map((item) => {
          const Icon = item.icon;
          const isActive = location.pathname.startsWith(item.path);
          
          return (
            <NavLink
              key={item.path}
              to={item.path}
              className={`flex items-center gap-3 px-3 py-3 rounded-lg mb-1 transition-all duration-200 ${
                isActive 
                  ? 'bg-cyber-blue text-white' 
                  : 'text-gray-400 hover:bg-gray-700 hover:text-white'
              }`}
            >
              <Icon size={20} className="flex-shrink-0" />
              {isOpen && <span className="font-medium">{item.label}</span>}
            </NavLink>
          );
        })}
      </nav>

      {/* User Info */}
      {isOpen && user && (
        <div className="absolute bottom-0 left-0 right-0 p-4 border-t border-gray-700">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-full bg-gradient-to-br from-cyber-blue to-neon-green flex items-center justify-center text-white font-bold">
              {user.name?.[0]?.toUpperCase() || user.email?.[0]?.toUpperCase()}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-sm font-medium text-white truncate">{user.name || user.email}</p>
              <p className="text-xs text-gray-400 capitalize">{user.role}</p>
            </div>
          </div>
        </div>
      )}
    </aside>
  );
};

export default Sidebar;
