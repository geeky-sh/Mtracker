import React from 'react';
import { NavigationContainer } from '@react-navigation/native';
import { createBottomTabNavigator } from '@react-navigation/bottom-tabs';
import { Ionicons } from '@expo/vector-icons';
import { HomeScreen } from '@/screens/HomeScreen';
import { NewActivityScreen } from '@/screens/NewActivityScreen';
import { TrackActivityScreen } from '@/screens/TrackActivityScreen';

const Tab = createBottomTabNavigator();

export function AppNavigator() {
  return (
    <NavigationContainer>
      <Tab.Navigator
        screenOptions={({ route }) => ({
          tabBarIcon: ({ color }) => {
            const icons: Record<string, string> = {
              Home: 'analytics-outline',
              New: 'add-circle-outline',
              Track: 'checkmark-circle-outline',
            };
            return <Ionicons name={icons[route.name] as any} size={28} color={color} />;
          },
          tabBarActiveTintColor: '#1a1a1a',
          tabBarInactiveTintColor: '#555',
          tabBarIconStyle: { marginTop: 4 },
          tabBarLabelStyle: { fontSize: 13, fontWeight: '600' },
          tabBarStyle: {
            backgroundColor: '#fff',
            borderTopWidth: 1,
            borderTopColor: '#eee',
            height: 70,
            paddingBottom: 10,
          },
          headerStyle: { backgroundColor: '#f8f9fa', elevation: 0, shadowOpacity: 0 },
          headerShown: false,
        })}
      >
        <Tab.Screen name="Home" component={HomeScreen} />
        <Tab.Screen name="New" component={NewActivityScreen} />
        <Tab.Screen name="Track" component={TrackActivityScreen} />
      </Tab.Navigator>
    </NavigationContainer>
  );
}
